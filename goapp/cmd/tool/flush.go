package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hibiken/asynq"
)

const flushModes = "all|active|pending|scheduled|retry"

func runFlush(queues []string, inspector *asynq.Inspector, args []string) error {
	if len(args) == 1 && !strings.HasPrefix(args[0], "-") {
		if !isFlushMode(args[0]) {
			return fmt.Errorf("unknown flush mode: %s (use %s)", args[0], flushModes)
		}
		return flushByMode(queues, inspector, args[0])
	}

	fs := flag.NewFlagSet("flush", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	all := fs.Bool("all", false, "cancel or archive tasks in every queue state (default)")
	active := fs.Bool("active", false, "cancel active tasks")
	pending := fs.Bool("pending", false, "archive pending tasks")
	scheduled := fs.Bool("scheduled", false, "archive scheduled tasks")
	retry := fs.Bool("retry", false, "archive retry tasks")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if boolCount(*all, *active, *pending, *scheduled, *retry) > 1 {
		return fmt.Errorf("flush flags are mutually exclusive: use one of --all, --active, --pending, --scheduled, --retry")
	}
	mode := "all"
	for candidate, selected := range map[string]bool{
		"active": *active, "pending": *pending, "scheduled": *scheduled, "retry": *retry,
	} {
		if selected {
			mode = candidate
		}
	}
	return flushByMode(queues, inspector, mode)
}

func isFlushMode(mode string) bool {
	switch mode {
	case "all", "active", "pending", "scheduled", "retry":
		return true
	default:
		return false
	}
}

type flushCounts struct {
	active    int
	pending   int
	scheduled int
	retry     int
}

func (c flushCounts) total() int { return c.active + c.pending + c.scheduled + c.retry }

func flushByMode(queues []string, inspector *asynq.Inspector, mode string) error {
	existing, err := existingQueueSet(inspector)
	if err != nil {
		return err
	}
	var counts flushCounts
	if mode == "all" || mode == "active" {
		counts.active, err = cancelActive(queues, existing, inspector)
		if err != nil {
			return err
		}
	}
	archiveModes := []struct {
		name    string
		archive func(string) (int, error)
		count   *int
	}{
		{"pending", inspector.ArchiveAllPendingTasks, &counts.pending},
		{"scheduled", inspector.ArchiveAllScheduledTasks, &counts.scheduled},
		{"retry", inspector.ArchiveAllRetryTasks, &counts.retry},
	}
	for _, item := range archiveModes {
		if mode != "all" && mode != item.name {
			continue
		}
		*item.count, err = archiveTasks(queues, existing, item.archive)
		if err != nil {
			return err
		}
	}
	_, _ = fmt.Printf("flushed %s: active=%d pending=%d scheduled=%d retry=%d total=%d\n",
		mode, counts.active, counts.pending, counts.scheduled, counts.retry, counts.total())
	return nil
}

func cancelActive(queues []string, existing map[string]struct{}, inspector *asynq.Inspector) (int, error) {
	n := 0
	for _, queueName := range queues {
		if _, ok := existing[queueName]; !ok {
			continue
		}
		tasks, err := inspector.ListActiveTasks(queueName)
		if errors.Is(err, asynq.ErrQueueNotFound) {
			continue
		}
		if err != nil {
			return n, err
		}
		for _, task := range tasks {
			if err := inspector.CancelProcessing(task.ID); err != nil {
				return n, err
			}
			n++
		}
	}
	return n, nil
}

func archiveTasks(queues []string, existing map[string]struct{}, archive func(string) (int, error)) (int, error) {
	n := 0
	for _, queueName := range queues {
		if _, ok := existing[queueName]; !ok {
			continue
		}
		count, err := archive(queueName)
		if errors.Is(err, asynq.ErrQueueNotFound) {
			continue
		}
		if err != nil {
			return n, err
		}
		n += count
	}
	return n, nil
}

func existingQueueSet(inspector *asynq.Inspector) (map[string]struct{}, error) {
	queues, err := inspector.Queues()
	if err != nil {
		return nil, err
	}
	set := make(map[string]struct{}, len(queues))
	for _, queueName := range queues {
		set[queueName] = struct{}{}
	}
	return set, nil
}

func boolCount(values ...bool) int {
	n := 0
	for _, value := range values {
		if value {
			n++
		}
	}
	return n
}
