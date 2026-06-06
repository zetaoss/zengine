package aieditjob

import (
	"context"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"
	"github.com/zetaoss/zengine/goapp/models"
)

const (
	nannyJobName          = "ai-edit-nanny"
	nannyJobTimeout       = 1 * time.Minute
	nannyJobRetryInterval = 30 * time.Second
	batchSize             = 100
)

type NannyJob struct{}

func NewNannyJob() *NannyJob {
	return &NannyJob{}
}

func (j *NannyJob) Name() string { return nannyJobName }

func (j *NannyJob) Timeout() time.Duration { return nannyJobTimeout }

func (j *NannyJob) Run(ctx context.Context, jobCtx job.JobContext, _ any) job.Result {
	db, err := jobCtx.GetDB()
	if err != nil {
		return job.Error(err)
	}

	now := time.Now()

	candidatePhases := []models.AIEditPhase{
		models.AIEditPhasePending,
		models.AIEditPhaseGenerating,
		models.AIEditPhasePublishing,
		models.AIEditPhaseRetrying,
	}
	candidates := make([]models.AIEdit, 0, batchSize)
	if err := db.WithContext(ctx).Table("aiedit_tasks").
		Where("phase IN ?", candidatePhases).
		Order("id").
		Limit(batchSize).
		Find(&candidates).Error; err != nil {
		return job.Error(err)
	}

	enqueued := 0
	for _, task := range candidates {
		if !isReadyToEnqueue(task, now) {
			continue
		}

		_, err := Enqueue(ctx, jobCtx, task.ID)
		if err != nil {
			return job.Error(err)
		}
		enqueued++
	}

	return job.Success(app.H{
		"enqueued":   enqueued,
		"candidates": len(candidates),
	})
}

func isReadyToEnqueue(task models.AIEdit, now time.Time) bool {
	if task.UpdatedAt.IsZero() {
		// If timestamp is invalid, probably best to try and recover it.
		return true
	}

	switch task.Phase {
	case models.AIEditPhasePending, models.AIEditPhaseGenerating, models.AIEditPhasePublishing:
		// Task is considered stale and needs recovery after jobTimeout.
		return now.After(task.UpdatedAt.Add(aiEditJobTimeout + 1*time.Minute))
	case models.AIEditPhaseRetrying:
		// Task is ready for its scheduled retry.
		return now.After(task.UpdatedAt.Add(nannyJobRetryInterval))
	}
	return false
}
