package runboxjob

import (
	"context"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"
)

type PrunerJob struct{}

const (
	prunerJobName    = "runbox-pruner"
	prunerJobTimeout = 1 * time.Minute
)

func NewPrunerJob() *PrunerJob {
	return &PrunerJob{}
}

func (j *PrunerJob) Name() string { return prunerJobName }

func (j *PrunerJob) Timeout() time.Duration { return prunerJobTimeout }

func (j *PrunerJob) Run(ctx context.Context, jobCtx job.JobContext, _ any) job.Result {
	db, err := jobCtx.GetDB()
	if err != nil {
		return job.Error(err)
	}

	now := time.Now()
	res := db.WithContext(ctx).Table("runboxes").
		Where("phase = ?", "pending").
		Updates(app.H{
			"phase":      "failed",
			"updated_at": now,
		})
	if res.Error != nil {
		return job.Error(res.Error)
	}

	return job.Success(app.H{
		"failed_at": now.UTC().Format(time.RFC3339),
		"updated":   res.RowsAffected,
	})
}
