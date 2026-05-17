package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/worker"
	"github.com/zetaoss/zengine/goapp/worker/queue"
	"github.com/zetaoss/zengine/goapp/worker/registry"
	"github.com/zetaoss/zengine/goapp/worker/tablewriter"
)

const (
	defaultLimit = 200
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}
	config.ConfigureSlog(cfg.App.LogLevel)

	wkr, err := worker.New(cfg)
	if err != nil {
		slog.Error("failed to create worker", "err", err)
		os.Exit(1)
	}

	if err := run(wkr, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(wkr *worker.Worker, args []string) error {
	if len(args) == 0 {
		printUsage(wkr)
		return nil
	}

	switch args[0] {
	case "list":
		return runList(wkr, args[1:])
	case "flush":
		return runFlush(wkr, args[1:])
	case "help", "-h", "--help":
		printUsage(wkr)
		return nil
	default:
		return runJobCommand(wkr, args[0], args[1:])
	}
}

func runList(wkr *worker.Worker, args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
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

func runFlush(wkr *worker.Worker, args []string) error {
	if len(args) == 1 && !strings.HasPrefix(args[0], "-") {
		switch args[0] {
		case "all", "running", "pending":
			return flushByMode(wkr, args[0])
		default:
			return fmt.Errorf("unknown flush mode: %s (use all|running|pending)", args[0])
		}
	}

	fs := flag.NewFlagSet("flush", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	all := fs.Bool("all", false, "flush all jobs (default)")
	running := fs.Bool("running", false, "flush running jobs only")
	pending := fs.Bool("pending", false, "flush pending jobs only")
	if err := fs.Parse(args); err != nil {
		return err
	}

	selected := 0
	if *all {
		selected++
	}
	if *running {
		selected++
	}
	if *pending {
		selected++
	}
	if selected > 1 {
		return fmt.Errorf("flush flags are mutually exclusive: use one of --all, --running, --pending")
	}

	mode := "all"
	if *running {
		mode = "running"
	}
	if *pending {
		mode = "pending"
	}

	return flushByMode(wkr, mode)
}

func flushByMode(wkr *worker.Worker, mode string) error {
	reason := "flushed by ctl"
	switch mode {
	case "running":
		n, err := wkr.FlushRunning(reason)
		if err != nil {
			return err
		}
		_, _ = fmt.Printf("flushed running: %d\n", n)
	case "pending":
		n, err := wkr.FlushPending(reason)
		if err != nil {
			return err
		}
		_, _ = fmt.Printf("flushed pending: %d\n", n)
	default:
		rn, err := wkr.FlushRunning(reason)
		if err != nil {
			return err
		}
		pn, err := wkr.FlushPending(reason)
		if err != nil {
			return err
		}
		_, _ = fmt.Printf("flushed all: running=%d pending=%d total=%d\n", rn, pn, rn+pn)
	}
	return nil
}

func runJobCommand(wkr *worker.Worker, jobName string, args []string) error {
	known := map[string]struct{}{}
	for _, name := range wkr.JobNames() {
		known[name] = struct{}{}
	}
	if _, ok := known[jobName]; !ok {
		return fmt.Errorf("unknown command or job: %s", jobName)
	}

	var input []byte
	if len(args) > 0 {
		input = []byte(strings.Join(args, " "))
		var js any
		if err := json.Unmarshal(input, &js); err != nil {
			return fmt.Errorf("invalid JSON input for job %s: %w", jobName, err)
		}
	}

	output, err := wkr.RunJobWithResult(jobName, input)
	if err != nil {
		return err
	}
	if len(output) == 0 {
		_, _ = fmt.Println("ok")
		return nil
	}
	_, _ = fmt.Println(string(output))
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

func printUsage(wkr *worker.Worker) {
	_, _ = fmt.Println("usage:")
	_, _ = fmt.Println("  ctl <command> [options]")
	_, _ = fmt.Println()

	_, _ = fmt.Println("available commands:")
	_, _ = fmt.Printf("  %-30s %s\n", "list [--watch|-w]", "show job specs and queue state")
	_, _ = fmt.Printf("  %-30s %s\n", "flush (all|running|pending)", "remove queued/running jobs from queue storage")
	_, _ = fmt.Printf("  %-30s %s\n", "help", "show this help message")
	_, _ = fmt.Printf("  %-30s %s\n", "<job> [json-input]", "run a job immediately (direct run)")
	_, _ = fmt.Println()

	_, _ = fmt.Println("available jobs:")
	for _, name := range wkr.JobNames() {
		_, _ = fmt.Printf("  %-30s %s\n", name, "run "+name+" job immediately")
	}
}
