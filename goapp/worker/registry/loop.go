package registry

import (
	"context"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/loopq"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
)

// Looper is optionally implemented by tasks that run in a self-perpetuating loop.
type Looper interface {
	LoopInterval(ctx context.Context, tc taskctx.Context) (time.Duration, bool, error)
}

// LoopIntervalFunc returns the current loop interval duration and whether the loop is active.
type LoopIntervalFunc func(ctx context.Context, tc taskctx.Context) (time.Duration, bool, error)

// LoopConfig configures self-perpetuating dynamic-interval tasks.
type LoopConfig struct {
	Interval LoopIntervalFunc
}

// loop marks a task as a self-perpetuating loop. It disables retries (preventing loop duplication)
// and schedules subsequent runs automatically on completion.
// If an intervalFunc is omitted, the task must implement the Looper interface.
func loop(fn ...LoopIntervalFunc) option {
	return func(s *Spec) {
		s.MaxRetries = 0
		var intervalFn LoopIntervalFunc
		if len(fn) > 0 && fn[0] != nil {
			intervalFn = fn[0]
		}
		s.Loop = &LoopConfig{Interval: intervalFn}
	}
}

func loopQueue(spec Spec) string {
	if spec.Queue == "" {
		return "default"
	}
	return spec.Queue
}

func newLoopTask(spec Spec, delay time.Duration) (*asynq.Task, []asynq.Option) {
	return asynq.NewTask(spec.Type, nil), loopq.Options(loopQueue(spec), spec.Timeout, delay)
}

// scheduleNextLoop enqueues the next run of a finished loop task unless one is already waiting.
func scheduleNextLoop(ctx context.Context, tc taskctx.Context, spec Spec) {
	d, active, err := spec.Loop.Interval(ctx, tc)
	if err != nil || !active || d <= 0 {
		if err != nil {
			slog.Error("[registry] Failed to read loop interval", "type", spec.Type, "err", err)
		}
		return
	}
	task, opts := newLoopTask(spec, d)
	enqueued, err := loopq.ScheduleNext(ctx, tc.Config(), tc, loopQueue(spec), task, opts...)
	switch {
	case err != nil:
		slog.Error("[registry] Failed to schedule next loop task", "type", spec.Type, "err", err)
	case enqueued:
		slog.Info("[registry] Scheduled next loop task", "type", spec.Type, "in", d)
	}
}

// EnsureLoopTask enqueues a loop task when its interval is active and none is waiting or active.
// It bootstraps loops on worker start and restores lost ones from the hourly nanny.
func EnsureLoopTask(ctx context.Context, tc taskctx.Context, spec Spec) (bool, error) {
	if spec.Loop == nil || spec.Loop.Interval == nil {
		return false, nil
	}
	d, active, err := spec.Loop.Interval(ctx, tc)
	if err != nil || !active || d <= 0 {
		return false, err
	}
	task, opts := newLoopTask(spec, min(30*time.Second, d))
	return loopq.Ensure(ctx, tc.Config(), tc, loopQueue(spec), task, opts...)
}

// LoopNannyTask is the generic hourly watcher that verifies every loop task is running.
type LoopNannyTask struct {
	reg *Registry
}

func NewLoopNannyTask(r *Registry) *LoopNannyTask {
	return &LoopNannyTask{reg: r}
}

func (n *LoopNannyTask) Execute(ctx context.Context, tc taskctx.Context, _ any) (app.H, error) {
	restored := 0
	activeCount := 0

	for _, spec := range n.reg.AllSpecs() {
		if spec.Loop == nil || spec.Loop.Interval == nil {
			continue
		}
		_, active, err := spec.Loop.Interval(ctx, tc)
		if err == nil && active {
			activeCount++
		}

		enqueued, err := EnsureLoopTask(ctx, tc, spec)
		if err != nil {
			slog.Warn("[loop-nanny] Failed to ensure loop task", "type", spec.Type, "err", err)
		} else if enqueued {
			slog.Warn("[loop-nanny] Restored missing loop task", "type", spec.Type)
			restored++
		}
	}

	return app.H{
		"active_loops": activeCount,
		"restored":     restored,
	}, nil
}
