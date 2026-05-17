package scheduler

import (
	"context"
	"time"

	"github.com/zetaoss/zengine/goapp/app/job"
	"github.com/zetaoss/zengine/goapp/worker/registry"
)

type Scheduler struct {
	ctx      context.Context
	enqueuer job.Enqueuer
	jobs     []registry.Spec
	lastRun  map[string]string
}

func New(ctx context.Context, enqueuer job.Enqueuer, list []registry.Spec) *Scheduler {
	return &Scheduler{ctx: ctx, enqueuer: enqueuer, jobs: list, lastRun: map[string]string{}}
}

func (s *Scheduler) Run() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		now := time.Now()
		s.tick(now)
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Scheduler) tick(now time.Time) {
	for _, spec := range s.jobs {
		if spec.Schedule == nil {
			continue
		}
		name := spec.Job.Name()
		key, ok := spec.Schedule.Key(now)
		if !ok {
			continue
		}
		if s.lastRun[name] == key {
			continue
		}
		s.lastRun[name] = key
		_, _ = s.enqueuer.Enqueue(s.ctx, job.Request{
			JobName:    name,
			Queue:      spec.Queue,
			Input:      spec.Input,
			MaxRetries: spec.MaxRetries,
			DedupeKey:  name + ":" + key,
		})
	}
}
