package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hibiken/asynq"
	"github.com/zetaoss/zengine/goapp/app/appctx"
	"github.com/zetaoss/zengine/goapp/app/config"
	appredis "github.com/zetaoss/zengine/goapp/app/redis"
	"github.com/zetaoss/zengine/goapp/worker/registry"
)

func runTaskCommand(cfg *config.Config, reg *registry.Registry, taskType string, args []string) error {
	_, ok := reg.FindSpec(taskType)
	if !ok {
		return fmt.Errorf("unknown command or task: %s", taskType)
	}
	var raw []byte
	if len(args) > 0 {
		raw = []byte(strings.Join(args, " "))
		var value any
		if err := json.Unmarshal(raw, &value); err != nil {
			return fmt.Errorf("invalid JSON input for task %s: %w", taskType, err)
		}
	}
	taskCtx, err := appctx.NewAppContext(cfg)
	if err != nil {
		return err
	}
	conn, err := appredis.AsynqConnOpt(cfg)
	if err != nil {
		return err
	}
	client := asynq.NewClient(conn)
	taskCtx.TaskEnqueuer = client
	output, runErr := reg.RunDirect(context.Background(), taskCtx, taskType, raw)
	closeErr := client.Close()
	if runErr != nil {
		return runErr
	}
	if closeErr != nil {
		return closeErr
	}
	if len(output) == 0 {
		_, _ = fmt.Fprintln(os.Stdout, "ok")
		return nil
	}
	_, _ = fmt.Fprintln(os.Stdout, string(output))
	return nil
}

func printUsage(reg *registry.Registry) {
	_, _ = fmt.Println("usage:")
	_, _ = fmt.Println("  ctl <command> [options]")
	_, _ = fmt.Println()
	_, _ = fmt.Println("available commands:")
	_, _ = fmt.Printf("  %-30s %s\n", "migrate", "run database migrations")
	_, _ = fmt.Printf("  %-30s %s\n", "tasks [--watch|-w]", "show task specs and queue state")
	_, _ = fmt.Printf("  %-30s %s\n", "flush (all|active|pending|scheduled|retry)", "cancel or archive tasks by Asynq state")
	_, _ = fmt.Printf("  %-30s %s\n", "routes", "show api routes")
	_, _ = fmt.Printf("  %-30s %s\n", "help", "show this help message")
	_, _ = fmt.Printf("  %-30s %s\n", "<task> [json-input]", "run a task immediately (direct run)")
	_, _ = fmt.Println()
	_, _ = fmt.Println("available tasks:")
	for _, name := range reg.Names() {
		_, _ = fmt.Printf("  %-30s %s\n", name, "run "+name+" task immediately")
	}
}
