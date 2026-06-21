package aieditjob

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
	"github.com/zetaoss/zengine/goapp/services/llmsvc"

	"gorm.io/gorm"
)

type AIEditJob struct{}

const (
	aiEditJobName    = "ai-edit"
	aiEditJobTimeout = 10 * time.Minute
	aiEditJobQueue   = "default"
)

func NewAIEditJob() *AIEditJob {
	return &AIEditJob{}
}

func Enqueue(ctx context.Context, jobCtx job.JobContext, taskID int) (uint64, error) {
	return (&AIEditJob{}).Enqueue(ctx, jobCtx, taskID)
}

type aiEditPayload struct {
	TaskID int `json:"task_id"`
}

func (j *AIEditJob) Name() string { return aiEditJobName }

func (j *AIEditJob) Timeout() time.Duration { return aiEditJobTimeout }

func (j *AIEditJob) Enqueue(ctx context.Context, jobCtx job.JobContext, taskID int) (uint64, error) {
	return jobCtx.Enqueue(ctx, job.Request{
		JobName: aiEditJobName,
		Queue:   aiEditJobQueue,
		Input:   map[string]int{"task_id": taskID},
	})
}

func (j *AIEditJob) Run(ctx context.Context, jobCtx job.JobContext, p aiEditPayload) job.Result {
	db, err := jobCtx.GetDB()
	if err != nil {
		return job.Error(err)
	}

	now := time.Now()
	if p.TaskID < 1 {
		return job.Error(fmt.Errorf("AI Edit task_id is required"))
	}
	task, err := findRunnableTaskByID(db, p.TaskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return job.Success(app.H{"status": "idle"})
		}
		return job.Error(err)
	}
	slog.Info("[aiedit] Job picked", "task_id", task.ID, "page_id", task.PageID, "page_title", task.PageTitle, "phase", task.Phase)

	llmOutput := task.LLMOutput
	var actualModel string
	var llmInput string

	updates := app.H{
		"phase":      models.AIEditPhaseGenerating,
		"attempts":   task.Attempts + 1,
		"updated_at": now,
	}
	res := db.WithContext(ctx).Table("aiedit_tasks").
		Where("id = ? AND phase = ?", task.ID, task.Phase).
		Updates(updates)
	if res.Error != nil {
		return job.Error(res.Error)
	}
	if res.RowsAffected == 0 {
		return job.Error(fmt.Errorf("task #%d already claimed by another worker", task.ID))
	}

	if strings.TrimSpace(llmOutput) == "" {
		slog.Info("[aiedit] generating content", "task_id", task.ID)
		var genErr error
		llmOutput, actualModel, llmInput, genErr = j.generateContent(ctx, jobCtx, task)
		if genErr != nil {
			_ = db.WithContext(ctx).Table("aiedit_tasks").Where("id = ?", task.ID).Updates(app.H{
				"phase":       models.AIEditPhaseRetrying,
				"error_count": task.ErrorCount + 1,
				"last_error":  genErr.Error(),
				"llm_input":   llmInput,
				"llm_model":   actualModel,
				"updated_at":  time.Now(),
			}).Error
			return job.Error(genErr)
		}
	} else {
		slog.Info("[aiedit] using existing LLM output, skipping generation", "task_id", task.ID)
		llmInput = task.LLMInput
		actualModel = task.LLMModel
	}

	slog.Info("[aiedit] LLM generation completed", "task_id", task.ID, "model", actualModel)

	finalUpdates := app.H{
		"phase":      models.AIEditPhaseCompleted,
		"llm_input":  llmInput,
		"llm_model":  actualModel,
		"llm_output": llmOutput,
		"last_error": nil,
		"updated_at": time.Now(),
	}
	if err := db.WithContext(ctx).Table("aiedit_tasks").Where("id = ?", task.ID).Updates(finalUpdates).Error; err != nil {
		return job.Error(err)
	}

	return job.Success(app.H{
		"task_id": task.ID,
		"phase":   models.AIEditPhaseCompleted,
	})
}

func findRunnableTaskByID(gdb *gorm.DB, taskID int) (models.AIEdit, error) {
	var task models.AIEdit
	if err := gdb.Table("aiedit_tasks").Where("id = ?", taskID).Take(&task).Error; err != nil {
		return models.AIEdit{}, err
	}
	return task, nil
}

func (j *AIEditJob) generateContent(ctx context.Context, jobCtx job.JobContext, task models.AIEdit) (string, string, string, error) {
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
func CalculateRetryAt(task models.AIEdit) *string {
	if task.Phase != models.AIEditPhaseRetrying {
		return nil
	}
	if task.UpdatedAt.IsZero() {
		return nil
	}
	delay := time.Duration(float64(aiEditJobTimeout) * math.Pow(1.1, float64(max(0, task.ErrorCount-1))))
	retryAt := task.UpdatedAt.Add(delay).Format(time.RFC3339)
	return &retryAt
}
