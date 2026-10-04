package pingredis

import (
	"context"
	"fmt"
	"strings"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/config"
	appredis "github.com/zetaoss/zengine/goapp/app/redis"
	"github.com/zetaoss/zengine/goapp/app/taskctx"

	goredis "github.com/redis/go-redis/v9"
)

type PingRedisTask struct{}

func NewPingRedisTask() *PingRedisTask {
	return &PingRedisTask{}
}

func (j *PingRedisTask) Execute(ctx context.Context, taskCtx taskctx.Context, _ any) (app.H, error) {
	persist, err := pingVersion(ctx, appredis.OpenPersist, taskCtx)
	if err != nil {
		return nil, fmt.Errorf("persist redis: %w", err)
	}
	volatile, err := pingVersion(ctx, appredis.OpenVolatile, taskCtx)
	if err != nil {
		return nil, fmt.Errorf("volatile redis: %w", err)
	}

	return app.H{
		"target":   "redis",
		"message":  "pong",
		"version":  persist,
		"persist":  persist,
		"volatile": volatile,
	}, nil
}

func pingVersion(ctx context.Context, openFn func(*config.Config) (*goredis.Client, error), taskCtx taskctx.Context) (string, error) {
	client, err := openFn(taskCtx.Config())
	if err != nil {
		return "", err
	}
	defer client.Close()

	if err := client.Ping(ctx).Err(); err != nil {
		return "", err
	}

	if info, err := client.Info(ctx, "server").Result(); err == nil {
		for _, line := range strings.Split(info, "\n") {
			if strings.HasPrefix(line, "redis_version:") {
				return strings.TrimSpace(strings.TrimPrefix(line, "redis_version:")), nil
			}
		}
	}
	return "unknown", nil
}
