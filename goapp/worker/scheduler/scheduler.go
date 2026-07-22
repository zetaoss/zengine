package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"
	appredis "github.com/zetaoss/zengine/goapp/app/redis"
	"github.com/zetaoss/zengine/goapp/worker/registry"

	"github.com/hibiken/asynq"
)

type Scheduler struct {
	scheduler *asynq.Scheduler
	specs     []registry.Spec
}

func New(cfg *config.Config) (*Scheduler, error) {
	conn, err := appredis.AsynqConnOpt(cfg)
	if err != nil {
		return nil, fmt.Errorf("configure asynq redis: %w", err)
	}
	reg := registry.New()
	s := &Scheduler{
		scheduler: asynq.NewScheduler(conn, &asynq.SchedulerOpts{
			Location: time.Local,
			PostEnqueueFunc: func(info *asynq.TaskInfo, err error) {
				if err != nil {
					slog.Error("scheduled task enqueue failed", "err", err)
				}
			},
		}),
		specs: reg.AllSpecs(),
	}
	if err := s.register(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Scheduler) register() error {
	for _, spec := range s.specs {
		if spec.Cron == "" {
			continue
		}
		task, opts, err := registry.NewTask(spec, spec.Input)
		if err != nil {
			return err
		}
		if _, err = s.scheduler.Register(spec.Cron, task, opts...); err != nil {
			return fmt.Errorf("register schedule for %s: %w", spec.Type, err)
		}
	}
	return nil
}

func (s *Scheduler) Run(ctx context.Context) error {
	if err := s.scheduler.Start(); err != nil {
		return err
	}
	<-ctx.Done()
	s.scheduler.Shutdown()
	return nil
}
