package pingdb

import (
	"context"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
)

type PingDBTask struct{}

func NewPingDBTask() *PingDBTask {
	return &PingDBTask{}
}

func (j *PingDBTask) Execute(ctx context.Context, taskCtx taskctx.Context, _ any) (app.H, error) {
	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}

	var version string
	if err := db.WithContext(ctx).Raw("SELECT VERSION()").Scan(&version).Error; err != nil {
		version = "unknown"
	}

	return app.H{
		"target":  "database",
		"message": "pong",
		"version": version,
	}, nil
}
