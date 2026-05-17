package appctx

import (
	"context"
	"sync"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/app/database"
	"github.com/zetaoss/zengine/goapp/app/job"

	"gorm.io/gorm"
)

type AppContext struct {
	Cfg         *config.Config
	JobEnqueuer job.Enqueuer
	db          *gorm.DB
	dbOnce      sync.Once
	dbErr       error
}

func NewAppContext(cfg *config.Config) (*AppContext, error) {
	return &AppContext{
		Cfg: cfg,
	}, nil
}

func (c *AppContext) Config() *config.Config {
	return c.Cfg
}

func (c *AppContext) Enqueue(ctx context.Context, req job.Request) (uint64, error) {
	if c.JobEnqueuer == nil {
		return 0, nil
	}
	return c.JobEnqueuer.Enqueue(ctx, req)
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
