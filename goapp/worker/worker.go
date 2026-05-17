package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/zetaoss/zengine/goapp/app/appctx"
	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/app/job"
	"github.com/zetaoss/zengine/goapp/app/redis"
	"github.com/zetaoss/zengine/goapp/worker/queue"
	"github.com/zetaoss/zengine/goapp/worker/registry"
	"github.com/zetaoss/zengine/goapp/worker/scheduler"
)

type Worker struct {
	appCtx    *appctx.AppContext
	ctx       context.Context
	cancel    context.CancelFunc
	registry  *registry.Registry
	queue     *queue.RedisQueue
	scheduler *scheduler.Scheduler
	specs     []registry.Spec
	queues    []string
}

func New(cfg *config.Config) (*Worker, error) {
	appCtx, err := appctx.NewAppContext(cfg)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())

	redisClient, err := redis.Open(cfg)
	if err != nil {
		cancel()
		return nil, err
	}
	reg := registry.New()
	specs := reg.AllSpecs()
	q := queue.New(redisClient)
	appCtx.JobEnqueuer = q
	s := scheduler.New(ctx, q, specs)
	queues := registry.ConsumeQueues(specs)

	return &Worker{
		appCtx:    appCtx,
		ctx:       ctx,
		cancel:    cancel,
		registry:  reg,
		queue:     q,
		scheduler: s,
		specs:     specs,
		queues:    queues,
	}, nil
}

func (w *Worker) Run() {
	go w.scheduler.Run()
	for _, queue := range w.queues {
		go w.consumeQueueLoop(queue)
	}
	<-w.ctx.Done()
}

func (w *Worker) consumeQueueLoop(queue string) {
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
		}

		row, err := w.queue.ClaimPending(queue)
		if err != nil {
			slog.Error("worker queue claim error", "queue", queue, "err", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if row == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		_, runErr := w.executeJob(w.ctx, row.JobName, row.Payload)
		if runErr == nil {
			if err := w.queue.MarkSucceeded(row.ID); err != nil {
				slog.Error("worker queue mark success failed", "id", row.ID, "err", err)
			}
			continue
		}
		if err := w.queue.MarkRetryOrFailed(row, runErr); err != nil {
			slog.Error("worker queue mark retry/failed error", "id", row.ID, "err", err)
		}
	}
}

func (w *Worker) RunJob(name string) error {
	_, err := w.RunJobWithResult(name, nil)
	return err
}

func (w *Worker) RunJobWithInput(name string, input []byte) error {
	_, err := w.RunJobWithResult(name, input)
	return err
}

func (w *Worker) RunJobWithResult(name string, input []byte) ([]byte, error) {
	runID, err := w.queue.StartDirectRun(name)
	if err != nil {
		return nil, err
	}
	output, err := w.executeJob(w.ctx, name, input)
	if err != nil {
		_ = w.queue.MarkFailed(runID, err)
		return nil, err
	}
	_ = w.queue.MarkSucceeded(runID)
	return output, nil
}

func (w *Worker) JobNames() []string {
	return w.registry.Names()
}

func (w *Worker) Executors() []job.Job {
	return w.registry.Jobs()
}

func (w *Worker) Specs() []registry.Spec {
	return w.specs
}

func (w *Worker) RunningJobs(limit int) ([]queue.RunningJob, error) {
	return w.queue.ListRunning(limit)
}

func (w *Worker) PendingJobs(queue string, limit int) ([]queue.PendingJob, error) {
	return w.queue.ListPending(queue, limit)
}

func (w *Worker) Stop() {
	w.cancel()
}

func (w *Worker) FlushRunning(reason string) (int64, error) {
	return w.queue.FlushRunning(reason)
}

func (w *Worker) FlushPending(reason string) (int64, error) {
	return w.queue.FlushPending(reason)
}

func (w *Worker) Done() <-chan struct{} {
	return w.ctx.Done()
}

func (w *Worker) executeJob(ctx context.Context, name string, rawInput []byte) ([]byte, error) {
	jobItem, ok := w.registry.Find(name)
	if !ok {
		return nil, fmt.Errorf("unknown job: %s", name)
	}
	input, err := jobItem.Decode(rawInput)
	if err != nil {
		return nil, err
	}
	return w.registry.Run(ctx, w.appCtx, job.Request{JobName: name, Input: input})
}
