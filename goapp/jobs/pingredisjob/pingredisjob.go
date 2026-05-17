package pingredisjob

import (
	"context"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"
	appredis "github.com/zetaoss/zengine/goapp/app/redis"
)

type PingRedisJob struct{}

func NewPingRedisJob() *PingRedisJob {
	return &PingRedisJob{}
}

func (j *PingRedisJob) Name() string { return "ping-redis" }

func (j *PingRedisJob) Timeout() time.Duration { return 5 * time.Second }

func (j *PingRedisJob) Run(ctx context.Context, jobCtx job.JobContext, _ any) job.Result {
	client, err := appredis.Open(jobCtx.Config())
	if err != nil {
		return job.Error(err)
	}

	if err := client.Ping(ctx).Err(); err != nil {
		return job.Error(err)
	}

	var version string
	if info, err := client.Info(ctx, "server").Result(); err == nil {
		for _, line := range strings.Split(info, "\n") {
			if strings.HasPrefix(line, "redis_version:") {
				version = strings.TrimSpace(strings.TrimSpace(strings.TrimPrefix(line, "redis_version:")))
				break
			}
		}
	}
	if version == "" {
		version = "unknown"
	}

	return job.Success(app.H{
		"target":  "redis",
		"message": "pong",
		"version": version,
	})
}
