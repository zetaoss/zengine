package writerequestjob

import (
	"context"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"
)

type PrunerJob struct{}

const (
	prunerJobName    = "request-pruner"
	prunerJobTimeout = 5 * time.Minute
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

	res := db.WithContext(ctx).Table("write_requests").
		Where("is_matched = ? AND updated_at < ?", false, time.Now().AddDate(0, 0, -30)).
		Delete(nil)
	if res.Error != nil {
		return job.Error(res.Error)
	}

	return job.Success(app.H{
		"deleted": res.RowsAffected,
	})
}
