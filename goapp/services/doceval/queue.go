package doceval

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/app/loopq"
	appredis "github.com/zetaoss/zengine/goapp/app/redis"
)

// The evaluation and definition loops share their own queue, so listing and replacing their tasks never
// scans other tasks.
const (
	LoopTaskType       = "doc-eval-loop"
	DefineLoopTaskType = "doc-define-loop"
	LoopQueue          = "doc-eval"
	LoopTimeout        = 5 * time.Minute
	DefineLoopTimeout  = 2 * time.Minute
)

func loopTimeout(taskType string) time.Duration {
	if taskType == DefineLoopTaskType {
		return DefineLoopTimeout
	}
	return LoopTimeout
}

// ScheduleLoop makes the only waiting task of a loop (LoopTaskType or DefineLoopTaskType) a run after delay.
func ScheduleLoop(ctx context.Context, cfg *config.Config, enq loopq.Enqueuer, taskType string, delay time.Duration) error {
	_, err := loopq.Replace(ctx, cfg, enq, LoopQueue, asynq.NewTask(taskType, nil), loopq.Options(LoopQueue, loopTimeout(taskType), delay)...)
	return err
}

// StopLoop removes the waiting task of a loop; a running one finishes and schedules nothing more.
func StopLoop(ctx context.Context, cfg *config.Config, taskType string) error {
	return loopq.Clear(ctx, cfg, LoopQueue, taskType)
}

// RunNow makes the loop's only waiting task an immediate run, forced even when the interval is off,
// so the next scheduled run starts counting from it. If pageID > 0 or pageTitle is non-empty, that
// page is evaluated.
func RunNow(ctx context.Context, cfg *config.Config, enq loopq.Enqueuer, pageID int, pageTitle string) (*asynq.TaskInfo, error) {
	return loopq.Replace(ctx, cfg, enq, LoopQueue, runNowTask(pageID, pageTitle), loopq.Options(LoopQueue, LoopTimeout, 0)...)
}

func runNowTask(pageID int, pageTitle string) *asynq.Task {
	p := map[string]any{"force": true}
	if pageID > 0 {
		p["page_id"] = pageID
	}
	if title := strings.TrimSpace(pageTitle); title != "" {
		p["page_title"] = title
	}
	payload, _ := json.Marshal(p)
	return asynq.NewTask(LoopTaskType, payload)
}

// QueuedTask is a doc-eval-loop task waiting or running in asynq.
type QueuedTask struct {
	ID            string    `json:"id"`
	State         string    `json:"state"`
	PageID        int       `json:"page_id,omitempty"`
	PageTitle     string    `json:"page_title,omitempty"`
	NextProcessAt time.Time `json:"next_process_at,omitzero"`
	FailedAt      time.Time `json:"failed_at,omitzero"`
	LastError     string    `json:"last_error,omitempty"`
}

// Failed (archived) tasks are listed only from the last day, and only the most recent few.
const (
	maxFailedShown = 3
	failedShownFor = 24 * time.Hour
)

// QueuedLoopTasks lists the tasks of a loop (LoopTaskType or DefineLoopTaskType) in the queue: active,
// pending, and scheduled by next run time, then the most recent failed ones. The task is not retried, so a
// failure is archived by asynq and the loop goes on with its next scheduled run.
func QueuedLoopTasks(cfg *config.Config, taskType string) ([]QueuedTask, error) {
	conn, err := appredis.AsynqConnOpt(cfg)
	if err != nil {
		return nil, err
	}
	inspector := asynq.NewInspector(conn)
	defer func() { _ = inspector.Close() }()

	lists := []struct {
		state string
		list  func(string, ...asynq.ListOption) ([]*asynq.TaskInfo, error)
	}{
		{"active", inspector.ListActiveTasks},
		{"pending", inspector.ListPendingTasks},
		{"scheduled", inspector.ListScheduledTasks},
	}
	var out []QueuedTask
	for _, l := range lists {
		var group []QueuedTask
		err := loopq.ForEach(l.list, LoopQueue, func(t *asynq.TaskInfo) (bool, error) {
			if t.Type == taskType {
				var p struct {
					PageID    int    `json:"page_id"`
					PageTitle string `json:"page_title"`
				}
				_ = json.Unmarshal(t.Payload, &p)
				group = append(group, QueuedTask{
					ID:            t.ID,
					State:         l.state,
					PageID:        p.PageID,
					PageTitle:     p.PageTitle,
					NextProcessAt: t.NextProcessAt,
					LastError:     t.LastErr,
				})
			}
			return false, nil
		})
		if err != nil {
			return nil, err
		}
		slices.SortFunc(group, func(a, b QueuedTask) int { return a.NextProcessAt.Compare(b.NextProcessAt) })
		out = append(out, group...)
	}

	var failed []QueuedTask
	err = loopq.ForEach(inspector.ListArchivedTasks, LoopQueue, func(t *asynq.TaskInfo) (bool, error) {
		if t.Type == taskType && time.Since(t.LastFailedAt) < failedShownFor {
			var p struct {
				PageID    int    `json:"page_id"`
				PageTitle string `json:"page_title"`
			}
			_ = json.Unmarshal(t.Payload, &p)
			failed = append(failed, QueuedTask{
				ID:        t.ID,
				State:     "failed",
				PageID:    p.PageID,
				PageTitle: p.PageTitle,
				FailedAt:  t.LastFailedAt,
				LastError: t.LastErr,
			})
		}
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	slices.SortFunc(failed, func(a, b QueuedTask) int { return b.FailedAt.Compare(a.FailedAt) })
	return append(out, failed[:min(len(failed), maxFailedShown)]...), nil
}
