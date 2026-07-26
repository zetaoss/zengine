package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/cmd/util/tablewriter"
	"github.com/zetaoss/zengine/goapp/worker/registry"

	"github.com/hibiken/asynq"
)

const defaultLimit = 200

func runTasks(reg *registry.Registry, inspector *asynq.Inspector, args []string) error {
	fs := flag.NewFlagSet("tasks", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	watch := fs.Bool("watch", false, "watch and refresh")
	watchShort := fs.Bool("w", false, "watch and refresh")
	if err := fs.Parse(args); err != nil {
		return err
	}

	show := func() error {
		if *watch || *watchShort {
			fmt.Print("\033[H\033[J")
			_, _ = fmt.Printf("%s\n\n", time.Now().Format(time.RFC3339))
		}
		if err := printSpecs(reg); err != nil {
			return err
		}
		fmt.Println()
		return printActive(registry.ConsumeQueues(reg.AllSpecs()), inspector)
	}
	if err := show(); err != nil {
		return err
	}
	if !*watch && !*watchShort {
		return nil
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := show(); err != nil {
			return err
		}
	}
	return nil
}

func printSpecs(reg *registry.Registry) error {
	specs := reg.AllSpecs()
	_, _ = fmt.Printf("tasks(%d):\n", len(specs))
	tw := tablewriter.New(os.Stdout, "task", "timeout", "retries", "queue", "schedule")
	if err := tw.Header(); err != nil {
		return err
	}
	for _, spec := range specs {
		timeout := spec.Timeout
		queueName := spec.Queue
		if queueName == "" {
			queueName = "default"
		}
		schedule := spec.Cron
		if schedule == "" {
			schedule = "-"
		}
		if err := tw.Row(spec.Type, formatDuration(timeout), spec.MaxRetries, queueName, schedule); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func printActive(queues []string, inspector *asynq.Inspector) error {
	type displayTask struct {
		id, name, state, queue, attempt, due, payload string
	}
	rows := make([]displayTask, 0)
	for _, queueName := range queues {
		lists := []struct {
			state string
			load  func() ([]*asynq.TaskInfo, error)
		}{
			{"active", func() ([]*asynq.TaskInfo, error) {
				return inspector.ListActiveTasks(queueName, asynq.PageSize(defaultLimit))
			}},
			{"pending", func() ([]*asynq.TaskInfo, error) {
				return inspector.ListPendingTasks(queueName, asynq.PageSize(defaultLimit))
			}},
			{"scheduled", func() ([]*asynq.TaskInfo, error) {
				return inspector.ListScheduledTasks(queueName, asynq.PageSize(defaultLimit))
			}},
			{"retry", func() ([]*asynq.TaskInfo, error) {
				return inspector.ListRetryTasks(queueName, asynq.PageSize(defaultLimit))
			}},
		}
		for _, list := range lists {
			tasks, err := list.load()
			if errors.Is(err, asynq.ErrQueueNotFound) {
				continue
			}
			if err != nil {
				return err
			}
			for _, task := range tasks {
				due := "-"
				if !task.NextProcessAt.IsZero() {
					due = formatDue(task.NextProcessAt)
				}
				rows = append(rows, displayTask{
					id: task.ID, name: task.Type, state: list.state, queue: queueName,
					attempt: fmt.Sprintf("%d/%d", task.Retried+1, task.MaxRetry+1),
					due:     due, payload: compactPayload(task.Payload),
				})
			}
		}
	}

	_, _ = fmt.Printf("active tasks(%d):\n", len(rows))
	if len(rows) == 0 {
		_, _ = fmt.Println("No active tasks found.")
		return nil
	}
	tw := tablewriter.New(os.Stdout, "id", "task", "status", "queue", "attempt", "due", "payload")
	if err := tw.Header(); err != nil {
		return err
	}
	for _, row := range rows {
		if err := tw.Row(row.id, row.name, row.state, row.queue, row.attempt, row.due, row.payload); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func formatDue(t time.Time) string {
	d := time.Until(t).Truncate(time.Second)
	if d <= 0 {
		return "0s"
	}
	return d.String()
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	if d%time.Minute == 0 {
		return fmt.Sprintf("%dm", int(d/time.Minute))
	}
	if d%time.Second == 0 {
		return d.Truncate(time.Second).String()
	}
	return d.String()
}

func compactPayload(payload []byte) string {
	if len(payload) == 0 || string(payload) == "null" {
		return "-"
	}
	s := strings.TrimSpace(strings.ReplaceAll(string(payload), "\n", " "))
	const maxLen = 80
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
