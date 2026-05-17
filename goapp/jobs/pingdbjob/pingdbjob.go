package pingdbjob

import (
	"context"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"
)

type PingDBJob struct{}

func NewPingDBJob() *PingDBJob {
	return &PingDBJob{}
}

func (j *PingDBJob) Name() string { return "ping-db" }

func (j *PingDBJob) Timeout() time.Duration { return 10 * time.Second }

func (j *PingDBJob) Run(ctx context.Context, jobCtx job.JobContext, _ any) job.Result {
	db, err := jobCtx.GetDB()
	if err != nil {
		return job.Error(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return job.Error(err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return job.Error(err)
	}

	var version string
	if err := db.WithContext(ctx).Raw("SELECT VERSION()").Scan(&version).Error; err != nil {
		version = "unknown"
	}

	return job.Success(app.H{
		"target":  "database",
		"message": "pong",
		"version": version,
	})
}
