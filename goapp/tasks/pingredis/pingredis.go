package pingredis

import (
	"context"
	"strings"

	"github.com/zetaoss/zengine/goapp/app"
	appredis "github.com/zetaoss/zengine/goapp/app/redis"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
)

type PingRedisTask struct{}

func NewPingRedisTask() *PingRedisTask {
	return &PingRedisTask{}
}

func (j *PingRedisTask) Execute(ctx context.Context, taskCtx taskctx.Context, _ any) (app.H, error) {
	client, err := appredis.Open(taskCtx.Config())
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
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

	return app.H{
		"target":  "redis",
		"message": "pong",
		"version": version,
	}, nil
}
