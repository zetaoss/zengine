package loopq

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/hibiken/asynq"
)

type fakeQueue struct {
	tasks      []*asynq.TaskInfo
	nextID     int
	enqueueErr error
	deleted    []string
	startOnDel map[string]bool
}

func (q *fakeQueue) list(state asynq.TaskState) ListFunc {
	return func(_ string, _ ...asynq.ListOption) ([]*asynq.TaskInfo, error) {
		var out []*asynq.TaskInfo
		for _, t := range q.tasks {
			if t.State == state {
				out = append(out, t)
			}
		}
		return out, nil
	}
}

func (q *fakeQueue) ListPendingTasks(queue string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error) {
	return q.list(asynq.TaskStatePending)(queue, opts...)
}

func (q *fakeQueue) ListScheduledTasks(queue string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error) {
	return q.list(asynq.TaskStateScheduled)(queue, opts...)
}

func (q *fakeQueue) ListActiveTasks(queue string, opts ...asynq.ListOption) ([]*asynq.TaskInfo, error) {
	return q.list(asynq.TaskStateActive)(queue, opts...)
}

func (q *fakeQueue) find(id string) *asynq.TaskInfo {
	for _, t := range q.tasks {
		if t.ID == id {
			return t
		}
	}
	return nil
}

func (q *fakeQueue) DeleteTask(_ string, id string) error {
	t := q.find(id)
	if t == nil {
		return asynq.ErrTaskNotFound
	}
	if q.startOnDel[id] {
		t.State = asynq.TaskStateActive
		return errors.New("cannot delete task in active state")
	}
	q.deleted = append(q.deleted, id)
	q.tasks = slices.DeleteFunc(q.tasks, func(t *asynq.TaskInfo) bool { return t.ID == id })
	return nil
}

func (q *fakeQueue) GetTaskInfo(_ string, id string) (*asynq.TaskInfo, error) {
	if t := q.find(id); t != nil {
		return t, nil
	}
	return nil, asynq.ErrTaskNotFound
}

func (q *fakeQueue) EnqueueTask(_ context.Context, task *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	if q.enqueueErr != nil {
		return nil, q.enqueueErr
	}
	q.nextID++
	info := &asynq.TaskInfo{ID: fmt.Sprintf("new%d", q.nextID), Type: task.Type(), State: asynq.TaskStateScheduled}
	q.tasks = append(q.tasks, info)
	return info, nil
}

func (q *fakeQueue) add(id, taskType string, state asynq.TaskState) {
	q.tasks = append(q.tasks, &asynq.TaskInfo{ID: id, Type: taskType, State: state})
}

func (q *fakeQueue) ids(taskType string) []string {
	var out []string
	for _, t := range q.tasks {
		if t.Type == taskType {
			out = append(out, t.ID)
		}
	}
	return out
}

func TestOptions(t *testing.T) {
	got := map[asynq.OptionType]any{}
	for _, o := range Options("doc-eval", 5*time.Minute, 30*time.Second) {
		got[o.Type()] = o.Value()
	}
	want := map[asynq.OptionType]any{
		asynq.QueueOpt:     "doc-eval",
		asynq.ProcessInOpt: 30 * time.Second,
		asynq.MaxRetryOpt:  0,
		asynq.TimeoutOpt:   5 * time.Minute,
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("option %v = %v, want %v", k, got[k], v)
		}
	}
}

func TestScheduleNextSkipsWhenWaiting(t *testing.T) {
	q := &fakeQueue{}
	q.add("cur", "loop", asynq.TaskStateActive)
	q.add("other", "other", asynq.TaskStateScheduled)
	task := asynq.NewTask("loop", nil)
	lists := []ListFunc{q.ListPendingTasks, q.ListScheduledTasks}

	enqueued, err := enqueueUnless(context.Background(), q, q, "q", task, nil, lists...)
	if err != nil || !enqueued {
		t.Fatalf("expected the next run to be enqueued beside the active one, got %v %v", enqueued, err)
	}
	enqueued, err = enqueueUnless(context.Background(), q, q, "q", task, nil, lists...)
	if err != nil || enqueued {
		t.Fatalf("expected no second waiting run, got %v %v", enqueued, err)
	}
}

func TestEnsureCountsActive(t *testing.T) {
	q := &fakeQueue{}
	q.add("cur", "loop", asynq.TaskStateActive)
	enqueued, err := enqueueUnless(context.Background(), q, q, "q", asynq.NewTask("loop", nil), nil, q.ListPendingTasks, q.ListScheduledTasks, q.ListActiveTasks)
	if err != nil || enqueued {
		t.Fatalf("expected an active run to count, got %v %v", enqueued, err)
	}
}

func TestReplaceKeepsOnlyNewTask(t *testing.T) {
	q := &fakeQueue{}
	q.add("cur", "loop", asynq.TaskStateActive)
	q.add("old1", "loop", asynq.TaskStateScheduled)
	q.add("old2", "loop", asynq.TaskStatePending)
	q.add("other", "other", asynq.TaskStateScheduled)

	info, err := replace(context.Background(), q, q, "q", asynq.NewTask("loop", nil), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := q.ids("loop"); !slices.Equal(got, []string{"cur", info.ID}) {
		t.Fatalf("loop tasks = %v, want active and %s", got, info.ID)
	}
	if got := q.ids("other"); len(got) != 1 {
		t.Fatalf("other task type was touched: %v", got)
	}
}

func TestReplaceDeletesNothingWhenEnqueueFails(t *testing.T) {
	q := &fakeQueue{enqueueErr: errors.New("redis down")}
	q.add("old", "loop", asynq.TaskStateScheduled)

	if _, err := replace(context.Background(), q, q, "q", asynq.NewTask("loop", nil), nil); err == nil {
		t.Fatal("expected the enqueue error")
	}
	if len(q.deleted) != 0 {
		t.Fatalf("deleted %v after a failed enqueue", q.deleted)
	}
}

func TestReplaceIgnoresTaskThatStarted(t *testing.T) {
	q := &fakeQueue{startOnDel: map[string]bool{"old": true}}
	q.add("old", "loop", asynq.TaskStateScheduled)

	if _, err := replace(context.Background(), q, q, "q", asynq.NewTask("loop", nil), nil); err != nil {
		t.Fatalf("a task that started running must not fail the replace: %v", err)
	}
}
