package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/zetaoss/zengine/goapp/worker"
)

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
		_, _ = fmt.Fprintln(os.Stdout, "ok")
		return nil
	}
	_, _ = fmt.Fprintln(os.Stdout, string(output))
	return nil
}

func printUsage(wkr *worker.Worker) {
	_, _ = fmt.Println("usage:")
	_, _ = fmt.Println("  ctl <command> [options]")
	_, _ = fmt.Println()

	_, _ = fmt.Println("available commands:")
	_, _ = fmt.Printf("  %-30s %s\n", "migrate", "run database migrations")
	_, _ = fmt.Printf("  %-30s %s\n", "jobs [--watch|-w]", "show job specs and queue state")
	_, _ = fmt.Printf("  %-30s %s\n", "flush (all|running|pending)", "remove queued/running jobs from queue storage")
	_, _ = fmt.Printf("  %-30s %s\n", "routes", "show api routes")
	_, _ = fmt.Printf("  %-30s %s\n", "help", "show this help message")
	_, _ = fmt.Printf("  %-30s %s\n", "<job> [json-input]", "run a job immediately (direct run)")
	_, _ = fmt.Println()

	_, _ = fmt.Println("available jobs:")
	for _, name := range wkr.JobNames() {
		_, _ = fmt.Printf("  %-30s %s\n", name, "run "+name+" job immediately")
	}
}
