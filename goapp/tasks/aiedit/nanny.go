package aiedit

import (
	"context"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
	"github.com/zetaoss/zengine/goapp/models"
)

const (
	nannyTaskType          = "ai-edit-nanny"
	nannyTaskTimeout       = 1 * time.Minute
	nannyTaskRetryInterval = 30 * time.Second
	batchSize              = 100
)

type NannyTask struct{}

func NewNannyTask() *NannyTask {
	return &NannyTask{}
}

func (j *NannyTask) Execute(ctx context.Context, taskCtx taskctx.Context, _ any) (app.H, error) {
	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	candidatePhases := []models.AIEditPhase{
		models.AIEditPhasePending,
		models.AIEditPhaseGenerating,
		models.AIEditPhaseRetrying,
	}
	candidates := make([]models.AIEdit, 0, batchSize)
	if err := db.WithContext(ctx).Table("aiedit_tasks").
		Where("phase IN ?", candidatePhases).
		Order("id").
		Limit(batchSize).
		Find(&candidates).Error; err != nil {
		return nil, err
	}

	enqueued := 0
	for _, task := range candidates {
		if !isReadyToEnqueue(task, now) {
			continue
		}

		_, err := Enqueue(ctx, taskCtx, task.ID)
		if err != nil {
			return nil, err
		}
		enqueued++
	}

	return app.H{
		"enqueued":   enqueued,
		"candidates": len(candidates),
	}, nil
}

func isReadyToEnqueue(task models.AIEdit, now time.Time) bool {
	if task.UpdatedAt.IsZero() {
		// If timestamp is invalid, probably best to try and recover it.
		return true
	}

	switch task.Phase {
	case models.AIEditPhasePending, models.AIEditPhaseGenerating:
		// Task is considered stale and needs recovery after jobTimeout.
		return now.After(task.UpdatedAt.Add(aiEditTaskTimeout + 1*time.Minute))
	case models.AIEditPhaseRetrying:
		// Task is ready for its scheduled retry.
		return now.After(task.UpdatedAt.Add(nannyTaskRetryInterval))
	}
	return false
}
