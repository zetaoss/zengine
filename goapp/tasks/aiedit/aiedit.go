package aiedit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/app/taskctx"

	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/services/llmsvc"

	"gorm.io/gorm"
)

type AIEditTask struct{}

const (
	aiEditTaskType    = "ai-edit"
	aiEditTaskTimeout = 10 * time.Minute
	aiEditTaskQueue   = "default"
)

func NewAIEditTask() *AIEditTask {
	return &AIEditTask{}
}

type aiEditPayload struct {
	TaskID int `json:"task_id"`
}

func Enqueue(ctx context.Context, taskCtx taskctx.Context, taskID int) (*asynq.TaskInfo, error) {
	payload, err := json.Marshal(aiEditPayload{TaskID: taskID})
	if err != nil {
		return nil, err
	}
	return taskCtx.EnqueueTask(ctx, asynq.NewTask(aiEditTaskType, payload), asynq.Queue(aiEditTaskQueue), asynq.MaxRetry(3), asynq.Timeout(aiEditTaskTimeout))
}

func (j *AIEditTask) Execute(ctx context.Context, taskCtx taskctx.Context, p aiEditPayload) (app.H, error) {
	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if p.TaskID < 1 {
		return nil, fmt.Errorf("AI Edit task_id is required")
	}
	task, err := findRunnableTaskByID(db, p.TaskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return app.H{"status": "idle"}, nil
		}
		return nil, err
	}
	slog.Info("[aiedit] Task picked", "task_id", task.ID, "page_id", task.PageID, "page_title", task.PageTitle, "phase", task.Phase)

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
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("task #%d already claimed by another worker", task.ID)
	}

	if strings.TrimSpace(llmOutput) == "" {
		slog.Info("[aiedit] generating content", "task_id", task.ID)
		var genErr error
		llmOutput, actualModel, llmInput, genErr = j.generateContent(ctx, taskCtx, task)
		if genErr != nil {
			_ = db.WithContext(ctx).Table("aiedit_tasks").Where("id = ?", task.ID).Updates(app.H{
				"phase":       models.AIEditPhaseRetrying,
				"error_count": task.ErrorCount + 1,
				"last_error":  genErr.Error(),
				"llm_input":   llmInput,
				"llm_model":   actualModel,
				"updated_at":  time.Now(),
			}).Error
			return nil, genErr
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
		return nil, err
	}

	return app.H{
		"task_id": task.ID,
		"phase":   models.AIEditPhaseCompleted,
	}, nil
}

func findRunnableTaskByID(gdb *gorm.DB, taskID int) (models.AIEdit, error) {
	var task models.AIEdit
	if err := gdb.Table("aiedit_tasks").Where("id = ?", taskID).Take(&task).Error; err != nil {
		return models.AIEdit{}, err
	}
	return task, nil
}

func (j *AIEditTask) generateContent(ctx context.Context, taskCtx taskctx.Context, task models.AIEdit) (string, string, string, error) {
	llmInput := strings.TrimSpace(task.LLMInput)
	if llmInput == "" {
		return "", "", "", fmt.Errorf("llm_input is required")
	}
	content, actualModel, err := callLLM(ctx, taskCtx.Config(), llmInput)
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
	delay := time.Duration(float64(aiEditTaskTimeout) * math.Pow(1.1, float64(max(0, task.ErrorCount-1))))
	retryAt := task.UpdatedAt.Add(delay).Format(time.RFC3339)
	return &retryAt
}
