package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/zetaoss/zengine/goapp/app/appctx"
	"github.com/zetaoss/zengine/goapp/app/config"
	appredis "github.com/zetaoss/zengine/goapp/app/redis"
	"github.com/zetaoss/zengine/goapp/worker/registry"

	"github.com/hibiken/asynq"
)

const shutdownTimeout = 30 * time.Second

type Worker struct {
	appCtx   *appctx.AppContext
	client   *asynq.Client
	registry *registry.Registry
	server   *asynq.Server
	mux      *asynq.ServeMux
}

func New(cfg *config.Config) (*Worker, error) {
	appCtx, err := appctx.NewAppContext(cfg)
	if err != nil {
		return nil, err
	}
	conn, err := appredis.AsynqConnOpt(cfg)
	if err != nil {
		return nil, fmt.Errorf("configure asynq redis: %w", err)
	}
	reg := registry.New()
	client := asynq.NewClient(conn)
	appCtx.TaskEnqueuer = client

	w := &Worker{appCtx: appCtx, client: client, registry: reg, mux: asynq.NewServeMux()}
	w.registerHandlers()
	queues := make(map[string]int)
	for _, queueName := range registry.ConsumeQueues(reg.AllSpecs()) {
		queues[queueName] = 1
	}
	w.server = asynq.NewServer(conn, asynq.Config{
		Concurrency:     1,
		Queues:          queues,
		ShutdownTimeout: shutdownTimeout,
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			retried, _ := asynq.GetRetryCount(ctx)
			maxRetry, _ := asynq.GetMaxRetry(ctx)
			slog.Error("task failed", "task", task.Type(), "retry", retried, "max_retry", maxRetry, "err", err)
		}),
	})
	return w, nil
}

func (w *Worker) registerHandlers() {
	for _, spec := range w.registry.AllSpecs() {
		name := spec.Type
		w.mux.HandleFunc(name, func(ctx context.Context, task *asynq.Task) error {
			return w.registry.Process(ctx, w.appCtx, task)
		})
	}
}

func (w *Worker) Run(ctx context.Context) error {
	if err := w.server.Start(w.mux); err != nil {
		return fmt.Errorf("start asynq server: %w", err)
	}
	<-ctx.Done()
	w.Shutdown()
	return nil
}

func (w *Worker) Shutdown() {
	w.server.Shutdown()
}

func (w *Worker) Close() error { return w.client.Close() }
