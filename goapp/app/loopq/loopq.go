// Package loopq manages self-rescheduling loop tasks in asynq: a loop keeps at most one waiting task
// (pending or scheduled), and every change to that task happens under a Redis lock shared by all
// worker and server replicas.
package loopq

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	goredis "github.com/redis/go-redis/v9"
	"github.com/zetaoss/zengine/goapp/app/config"
	appredis "github.com/zetaoss/zengine/goapp/app/redis"
)

const (
	lockTTL  = 30 * time.Second
	lockWait = 10 * time.Second
	lockPoll = 100 * time.Millisecond
	pageSize = 100
)

// ErrBusy is returned when another replica holds the loop lock for longer than lockWait.
var ErrBusy = errors.New("loop task lock is busy")

var unlockScript = goredis.NewScript(`if redis.call("GET", KEYS[1]) == ARGV[1] then return redis.call("DEL", KEYS[1]) end return 0`)

// Enqueuer enqueues asynq tasks; taskctx.Context and the server context implement it.
type Enqueuer interface {
	EnqueueTask(context.Context, *asynq.Task, ...asynq.Option) (*asynq.TaskInfo, error)
}

// Inspector is the part of *asynq.Inspector that loopq uses.
type Inspector interface {
	ListPendingTasks(queue string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error)
	ListScheduledTasks(queue string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error)
	ListActiveTasks(queue string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error)
	DeleteTask(queue, id string) error
	GetTaskInfo(queue, id string) (*asynq.TaskInfo, error)
}

// ListFunc lists one state of a queue, such as (*asynq.Inspector).ListPendingTasks.
type ListFunc func(queue string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error)

// Options returns the enqueue options of a loop task. Loop tasks are never retried: a retry would
// run beside the next scheduled run.
func Options(queue string, timeout, delay time.Duration) []asynq.Option {
	return []asynq.Option{asynq.Queue(queue), asynq.ProcessIn(delay), asynq.MaxRetry(0), asynq.Timeout(timeout)}
}

// ForEach calls fn for every task of a queue state until fn returns stop or an error. A queue that
// does not exist yet has no tasks.
func ForEach(list ListFunc, queue string, fn func(*asynq.TaskInfo) (stop bool, err error)) error {
	for page := 1; ; page++ {
		tasks, err := list(queue, asynq.Page(page), asynq.PageSize(pageSize))
		if errors.Is(err, asynq.ErrQueueNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		for _, t := range tasks {
			stop, err := fn(t)
			if err != nil || stop {
				return err
			}
		}
		if len(tasks) < pageSize {
			return nil
		}
	}
}

// Ensure enqueues task unless a task of its type is waiting or active. It restores a loop that was
// lost, for example after an archived run.
func Ensure(ctx context.Context, cfg *config.Config, enq Enqueuer, queue string, task *asynq.Task, opts ...asynq.Option) (bool, error) {
	var enqueued bool
	err := withLock(ctx, cfg, queue, task.Type(), func(ins Inspector) (err error) {
		enqueued, err = enqueueUnless(ctx, ins, enq, queue, task, opts, ins.ListPendingTasks, ins.ListScheduledTasks, ins.ListActiveTasks)
		return err
	})
	return enqueued, err
}

// ScheduleNext enqueues the next run of a finishing loop task unless a task of its type is already
// waiting. The finishing task itself is active and is not counted.
func ScheduleNext(ctx context.Context, cfg *config.Config, enq Enqueuer, queue string, task *asynq.Task, opts ...asynq.Option) (bool, error) {
	var enqueued bool
	err := withLock(ctx, cfg, queue, task.Type(), func(ins Inspector) (err error) {
		enqueued, err = enqueueUnless(ctx, ins, enq, queue, task, opts, ins.ListPendingTasks, ins.ListScheduledTasks)
		return err
	})
	return enqueued, err
}

// Replace enqueues task and then deletes the other waiting tasks of its type, so the loop is never
// left without a waiting task: when the enqueue fails, nothing is deleted.
func Replace(ctx context.Context, cfg *config.Config, enq Enqueuer, queue string, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	var info *asynq.TaskInfo
	err := withLock(ctx, cfg, queue, task.Type(), func(ins Inspector) (err error) {
		info, err = replace(ctx, ins, enq, queue, task, opts)
		return err
	})
	return info, err
}

// Clear deletes the waiting tasks of taskType, stopping the loop after the active run, if any.
func Clear(ctx context.Context, cfg *config.Config, queue, taskType string) error {
	return withLock(ctx, cfg, queue, taskType, func(ins Inspector) error {
		return deleteWaiting(ins, queue, taskType, "")
	})
}

func enqueueUnless(ctx context.Context, ins Inspector, enq Enqueuer, queue string, task *asynq.Task, opts []asynq.Option, lists ...ListFunc) (bool, error) {
	for _, list := range lists {
		found, err := exists(list, queue, task.Type())
		if err != nil || found {
			return false, err
		}
	}
	if _, err := enq.EnqueueTask(ctx, task, opts...); err != nil {
		return false, err
	}
	return true, nil
}

func replace(ctx context.Context, ins Inspector, enq Enqueuer, queue string, task *asynq.Task, opts []asynq.Option) (*asynq.TaskInfo, error) {
	info, err := enq.EnqueueTask(ctx, task, opts...)
	if err != nil {
		return nil, err
	}
	if err := deleteWaiting(ins, queue, task.Type(), info.ID); err != nil {
		return info, fmt.Errorf("enqueued %s but failed to remove the previous one: %w", info.ID, err)
	}
	return info, nil
}

func exists(list ListFunc, queue, taskType string) (bool, error) {
	var found bool
	err := ForEach(list, queue, func(t *asynq.TaskInfo) (bool, error) {
		found = t.Type == taskType
		return found, nil
	})
	return found, err
}

func deleteWaiting(ins Inspector, queue, taskType, keepID string) error {
	var ids []string
	for _, list := range []ListFunc{ins.ListPendingTasks, ins.ListScheduledTasks} {
		err := ForEach(list, queue, func(t *asynq.TaskInfo) (bool, error) {
			if t.Type == taskType && t.ID != keepID {
				ids = append(ids, t.ID)
			}
			return false, nil
		})
		if err != nil {
			return err
		}
	}
	for _, id := range ids {
		if err := ins.DeleteTask(queue, id); err != nil && !startedOrGone(ins, queue, id, err) {
			return fmt.Errorf("delete task %s: %w", id, err)
		}
	}
	return nil
}

// startedOrGone reports whether a delete failed only because the task started running or finished
// between listing and deleting it. A running task schedules nothing while another one waits.
func startedOrGone(ins Inspector, queue, id string, err error) bool {
	if errors.Is(err, asynq.ErrTaskNotFound) {
		return true
	}
	info, infoErr := ins.GetTaskInfo(queue, id)
	return errors.Is(infoErr, asynq.ErrTaskNotFound) || (infoErr == nil && info.State != asynq.TaskStatePending && info.State != asynq.TaskStateScheduled)
}

// withLock runs fn while holding the loop lock of taskType in queue, waiting up to lockWait for it.
func withLock(ctx context.Context, cfg *config.Config, queue, taskType string, fn func(Inspector) error) error {
	rdb, err := appredis.OpenPersist(cfg)
	if err != nil {
		return fmt.Errorf("open redis for loop lock: %w", err)
	}
	defer func() { _ = rdb.Close() }()

	key := "lock:loop:" + queue + ":" + taskType
	token, err := newToken()
	if err != nil {
		return err
	}
	deadline := time.Now().Add(lockWait)
	for {
		acquired, err := rdb.SetNX(ctx, key, token, lockTTL).Result()
		if err != nil {
			return fmt.Errorf("acquire loop lock %s: %w", key, err)
		}
		if acquired {
			break
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%w: %s", ErrBusy, key)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(lockPoll):
		}
	}
	defer func() { _ = unlockScript.Run(context.Background(), rdb, []string{key}, token).Err() }()

	conn, err := appredis.AsynqConnOpt(cfg)
	if err != nil {
		return err
	}
	ins := asynq.NewInspector(conn)
	defer func() { _ = ins.Close() }()
	return fn(ins)
}

func newToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate loop lock token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
