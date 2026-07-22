package appctx

import (
	"context"
	"fmt"
	"sync"

	"github.com/hibiken/asynq"
	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/app/database"
	"github.com/zetaoss/zengine/goapp/app/taskctx"

	"gorm.io/gorm"
)

type AppContext struct {
	Cfg          *config.Config
	TaskEnqueuer taskctx.Enqueuer
	db           *gorm.DB
	dbOnce       sync.Once
	dbErr        error
}

func NewAppContext(cfg *config.Config) (*AppContext, error) {
	return &AppContext{
		Cfg: cfg,
	}, nil
}

func (c *AppContext) Config() *config.Config {
	return c.Cfg
}

func (c *AppContext) EnqueueTask(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	if c.TaskEnqueuer == nil {
		return nil, fmt.Errorf("task enqueuer is not configured")
	}
	return c.TaskEnqueuer.EnqueueContext(ctx, task, opts...)
}

func (c *AppContext) GetDB() (*gorm.DB, error) {
	if c.db != nil {
		return c.db, nil
	}
	c.dbOnce.Do(func() {
		if c.Cfg == nil {
			c.dbErr = gorm.ErrInvalidDB
			return
		}
		c.db, c.dbErr = database.Open(c.Cfg)
	})
	return c.db, c.dbErr
}
