package editbotjob

import (
	"context"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"
	"github.com/zetaoss/zengine/goapp/models"
)

const (
	nannyJobName          = "edit-bot-nanny"
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

	candidatePhases := []models.EditBotPhase{
		models.EditBotPhasePending,
		models.EditBotPhaseGenerating,
		models.EditBotPhasePublishing,
		models.EditBotPhaseRetryingGenerate,
		models.EditBotPhaseRetryingPublish,
	}
	candidates := make([]models.EditBot, 0, batchSize)
	if err := db.WithContext(ctx).Table("edit_tasks").
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

func isReadyToEnqueue(task models.EditBot, now time.Time) bool {
	updatedAt, err := time.Parse(time.RFC3339, task.UpdatedAt)
	if err != nil {
		// If timestamp is invalid, probably best to try and recover it.
		return true
	}

	switch task.Phase {
	case models.EditBotPhasePending, models.EditBotPhaseGenerating, models.EditBotPhasePublishing:
		// Task is considered stale and needs recovery after jobTimeout.
		return now.After(updatedAt.Add(editBotJobTimeout + 1*time.Minute))
	case models.EditBotPhaseRetryingGenerate, models.EditBotPhaseRetryingPublish:
		// Task is ready for its scheduled retry.
		return now.After(updatedAt.Add(nannyJobRetryInterval))
	}
	return false
}
