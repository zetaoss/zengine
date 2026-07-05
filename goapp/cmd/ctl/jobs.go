package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/cmd/ctl/tablewriter"
	"github.com/zetaoss/zengine/goapp/worker"
	"github.com/zetaoss/zengine/goapp/worker/queue"
	"github.com/zetaoss/zengine/goapp/worker/registry"
)

const defaultLimit = 200

func runJobs(wkr *worker.Worker, args []string) error {
	fs := flag.NewFlagSet("jobs", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	watch := fs.Bool("watch", false, "watch and refresh")
	watchShort := fs.Bool("w", false, "watch and refresh")
	if err := fs.Parse(args); err != nil {
		return err
	}

	show := func() error {
		isWatch := *watch || *watchShort
		if isWatch {
			// Watch-style redraw: move cursor to home and clear to end of screen.
			// This keeps the view anchored at the top while avoiding full scrollback churn.
			fmt.Print("\033[H\033[J")
			_, _ = fmt.Printf("%s\n\n", time.Now().Format(time.RFC3339))
		}
		if err := printSpecs(wkr); err != nil {
			return err
		}
		fmt.Println()
		if err := printRunning(wkr); err != nil {
			return err
		}
		return nil
	}

	if err := show(); err != nil {
		return err
	}

	if !*watch && !*watchShort {
		return nil
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if err := show(); err != nil {
			return err
		}
	}
	return nil
}

func printSpecs(wkr *worker.Worker) error {
	specs := wkr.Specs()
	_, _ = fmt.Printf("jobs(%d):\n", len(specs))
	tw := tablewriter.New(os.Stdout, "job", "timeout", "retries", "queue", "schedule")
	if err := tw.Header(); err != nil {
		return err
	}
	for _, spec := range specs {
		timeout := spec.Timeout
		if timeout == 0 {
			timeout = spec.Job.Timeout()
		}
		queueName := spec.Queue
		if queueName == "" {
			queueName = "default"
		}
		scheduleName := "-"
		if spec.Schedule != nil {
			scheduleName = spec.Schedule.Name()
		}
		if err := tw.Row(spec.Job.Name(), formatDuration(timeout), spec.MaxRetries, queueName, scheduleName); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func printRunning(wkr *worker.Worker) error {
	rows, err := wkr.RunningJobs(defaultLimit)
	if err != nil {
		return err
	}
	queues := collectQueues(wkr.Specs())
	pendingByQueue := make(map[string][]queue.PendingJob, len(queues))
	totalPending := 0
	for _, queueName := range queues {
		pendingRows, err := wkr.PendingJobs(queueName, defaultLimit)
		if err != nil {
			return err
		}
		pendingByQueue[queueName] = pendingRows
		totalPending += len(pendingRows)
	}
	_, _ = fmt.Printf("active jobs(%d):\n", len(rows)+totalPending)
	if len(rows) == 0 && totalPending == 0 {
		_, _ = fmt.Println("No active jobs found.")
		return nil
	}

	tw := tablewriter.New(os.Stdout, "id", "job", "status", "queue", "attempt", "age", "payload")
	if err := tw.Header(); err != nil {
		return err
	}
	for _, row := range rows {
		age := "-"
		if row.LockedAt != nil {
			age = time.Since(*row.LockedAt).Truncate(time.Second).String()
		}
		payload := compactPayload(row.Payload)
		if err := tw.Row(row.ID, row.JobName, "Running", row.Queue, row.Attempt, age, payload); err != nil {
			return err
		}
	}

	for _, queueName := range queues {
		pendingRows := pendingByQueue[queueName]
		for _, row := range pendingRows {
			age := "-"
			if row.RunAt != nil {
				wait := time.Until(*row.RunAt).Truncate(time.Second)
				if wait <= 0 {
					age = "0s"
				} else {
					age = wait.String()
				}
			}
			payload := compactPayload(row.Payload)
			if err := tw.Row(row.ID, row.JobName, "Pending", row.Queue, row.Attempt, age, payload); err != nil {
				return err
			}
		}
	}
	return tw.Flush()
}

func collectQueues(specs []registry.Spec) []string {
	seen := map[string]struct{}{"default": {}}
	for _, spec := range specs {
		q := spec.Queue
		if q == "" {
			q = "default"
		}
		seen[q] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for q := range seen {
		out = append(out, q)
	}
	sort.Strings(out)
	return out
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
	if len(payload) == 0 {
		return "-"
	}
	s := strings.ReplaceAll(string(payload), "\n", " ")
	s = strings.TrimSpace(s)
	const maxLen = 80
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
