package commonreportjob

import (
	"context"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"
	"github.com/zetaoss/zengine/goapp/models"
)

const (
	nannyJobName    = "common-report-nanny"
	nannyJobTimeout = 1 * time.Minute
	batchSize       = 100
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

	candidates := make([]models.CommonReport, 0, batchSize)
	if err := db.WithContext(ctx).Table("common_reports").
		Where("phase IN ?", []models.CommonReportPhase{models.CommonReportPhasePending, models.CommonReportPhaseRunning}).
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

func isReadyToEnqueue(task models.CommonReport, now time.Time) bool {
	updatedAt, err := time.Parse(time.RFC3339, task.UpdatedAt)
	if err != nil {
		return true
	}

	threshold := 1 * time.Minute // Default buffer for pending
	if task.Phase == models.CommonReportPhaseRunning {
		threshold = commonReportJobTimeout + 1*time.Minute
	}

	return now.After(updatedAt.Add(threshold))
}
