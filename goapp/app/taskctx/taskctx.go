package taskctx

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/zetaoss/zengine/goapp/app/config"
	"gorm.io/gorm"
)

// Context provides application dependencies used while processing a task.
type Context interface {
	GetDB() (*gorm.DB, error)
	Config() *config.Config
	EnqueueTask(context.Context, *asynq.Task, ...asynq.Option) (*asynq.TaskInfo, error)
}

// Enqueuer is implemented by an Asynq client.
type Enqueuer interface {
	EnqueueContext(context.Context, *asynq.Task, ...asynq.Option) (*asynq.TaskInfo, error)
}
