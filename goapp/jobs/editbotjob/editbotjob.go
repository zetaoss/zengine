package editbotjob

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/app/job"

	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/services/editbotsvc"
	"github.com/zetaoss/zengine/goapp/services/llmsvc"

	"gorm.io/gorm"
)

type EditBotJob struct{}

const (
	editBotJobName    = "edit-bot"
	editBotJobTimeout = 10 * time.Minute
	editBotJobQueue   = "default"
)

func NewEditBotJob() *EditBotJob {
	return &EditBotJob{}
}

func Enqueue(ctx context.Context, jobCtx job.JobContext, taskID int) (uint64, error) {
	return (&EditBotJob{}).Enqueue(ctx, jobCtx, taskID)
}

type editBotPayload struct {
	TaskID int `json:"task_id"`
}

func (j *EditBotJob) Name() string { return editBotJobName }

func (j *EditBotJob) Timeout() time.Duration { return editBotJobTimeout }

func (j *EditBotJob) Enqueue(ctx context.Context, jobCtx job.JobContext, taskID int) (uint64, error) {
	return jobCtx.Enqueue(ctx, job.Request{
		JobName: editBotJobName,
		Queue:   editBotJobQueue,
		Input:   map[string]int{"task_id": taskID},
	})
}

func (j *EditBotJob) Run(ctx context.Context, jobCtx job.JobContext, p editBotPayload) job.Result {
	db, err := jobCtx.GetDB()
	if err != nil {
		return job.Error(err)
	}

	now := time.Now()
	if p.TaskID < 1 {
		return job.Error(fmt.Errorf("edit-bot task_id is required"))
	}
	task, err := findRunnableTaskByID(db, p.TaskID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return job.Success(app.H{"status": "idle"})
		}
		return job.Error(err)
	}

	slog.Info("[editbot] Job picked", "task_id", task.ID, "title", task.Title, "phase", task.Phase)

	updates := app.H{
		"phase":      models.EditBotPhaseGenerating,
		"attempts":   task.Attempts + 1,
		"updated_at": now,
	}
	res := db.WithContext(ctx).Table("edit_tasks").
		Where("id = ? AND phase = ?", task.ID, task.Phase).
		Updates(updates)
	if res.Error != nil {
		return job.Error(res.Error)
	}
	if res.RowsAffected == 0 {
		return job.Error(fmt.Errorf("task #%d already claimed by another worker", task.ID))
	}

	slog.Info("[editbot] generating content", "task_id", task.ID)
	llmOutput, actualModel, llmInput, genErr := j.generateContent(ctx, jobCtx, task)
	if genErr != nil {
		_ = db.WithContext(ctx).Table("edit_tasks").Where("id = ?", task.ID).Updates(app.H{
			"phase":       models.EditBotPhaseRetryingGenerate,
			"error_count": task.ErrorCount + 1,
			"last_error":  genErr.Error(),
			"llm_input":   llmInput,
			"llm_model":   actualModel,
			"updated_at":  time.Now(),
		}).Error
		return job.Error(genErr)
	}

	slog.Info("[editbot] LLM generation success", "task_id", task.ID, "model", actualModel)

	_ = db.WithContext(ctx).Table("edit_tasks").Where("id = ?", task.ID).Updates(app.H{
		"llm_input":  llmInput,
		"llm_model":  actualModel,
		"llm_output": llmOutput,
		"updated_at": time.Now(),
	}).Error

	_ = db.WithContext(ctx).Table("edit_tasks").Where("id = ?", task.ID).Updates(app.H{
		"phase":      models.EditBotPhasePublishing,
		"updated_at": time.Now(),
	}).Error

	slog.Info("[editbot] publishing content", "task_id", task.ID)
	pubRes, pubErr := editbotsvc.PublishContent(jobCtx.Config(), task.Title, strings.TrimSpace(task.RequestType), llmOutput, task.ID)
	if pubErr != nil {
		newPhase := models.EditBotPhaseRetryingPublish
		var pErr *editbotsvc.PublishError
		if errors.As(pubErr, &pErr) {
			// Terminal errors that should not be retried
			switch pErr.Code {
			case "articleexists", "protectedpage", "permissiondenied", "nocreate-missing", "invalidtitle", "nosuchpageid":
				newPhase = models.EditBotPhaseRejected
			}
		}

		_ = db.WithContext(ctx).Table("edit_tasks").Where("id = ?", task.ID).Updates(app.H{
			"phase":       newPhase,
			"error_count": task.ErrorCount + 1,
			"last_error":  pubErr.Error(),
			"updated_at":  time.Now(),
		}).Error
		return job.Error(pubErr)
	}

	slog.Info("[editbot] publish completed", "task_id", task.ID, "revid", pubRes.Revid)

	finalUpdates := app.H{
		"phase":      models.EditBotPhaseCompleted,
		"revid":      pubRes.Revid,
		"last_error": nil,
		"updated_at": time.Now(),
	}
	if err := db.WithContext(ctx).Table("edit_tasks").Where("id = ?", task.ID).Updates(finalUpdates).Error; err != nil {
		return job.Error(err)
	}

	return job.Success(app.H{
		"task_id":   task.ID,
		"phase":     models.EditBotPhaseCompleted,
		"published": pubRes,
	})
}

func findRunnableTaskByID(gdb *gorm.DB, taskID int) (models.EditBot, error) {
	var task models.EditBot
	if err := gdb.Table("edit_tasks").Where("id = ?", taskID).Take(&task).Error; err != nil {
		return models.EditBot{}, err
	}
	return task, nil
}

func (j *EditBotJob) generateContent(ctx context.Context, jobCtx job.JobContext, task models.EditBot) (string, string, string, error) {
	llmInput := strings.TrimSpace(task.LLMInput)
	if llmInput == "" {
		return "", "", "", fmt.Errorf("llm_input is required")
	}
	content, actualModel, err := callLLM(ctx, jobCtx.Config(), llmInput)
	return content, actualModel, llmInput, err
}

func callLLM(ctx context.Context, cfg *config.Config, prompt string) (string, string, error) {
	svc := llmsvc.New(cfg)
	out, err := svc.Generate(ctx, llmsvc.Input{Prompt: prompt})
	if err != nil {
		return "", "", err
	}
	return out.Content, out.Model, nil
}

// CalculateRetryAt returns the next retry time for a task based on its error count.
func CalculateRetryAt(task models.EditBot) *string {
	if task.Phase != models.EditBotPhaseRetryingGenerate && task.Phase != models.EditBotPhaseRetryingPublish {
		return nil
	}
	t, err := time.Parse(time.RFC3339, task.UpdatedAt)
	if err != nil {
		return nil
	}
	delay := time.Duration(float64(editBotJobTimeout) * math.Pow(1.1, float64(max(0, task.ErrorCount-1))))
	retryAt := t.Add(delay).Format(time.RFC3339)
	return &retryAt
}
