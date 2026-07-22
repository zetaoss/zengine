package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/app/database"
	appredis "github.com/zetaoss/zengine/goapp/app/redis"
	"github.com/zetaoss/zengine/goapp/worker/registry"

	"github.com/hibiken/asynq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}
	config.ConfigureSlog(cfg.App.LogLevel)
	reg := registry.New()
	if err := run(cfg, reg, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(cfg *config.Config, reg *registry.Registry, args []string) error {
	if len(args) == 0 {
		printUsage(reg)
		return nil
	}
	switch args[0] {
	case "migrate":
		return database.RunMigrate(cfg)
	case "tasks":
		return withInspector(cfg, func(inspector *asynq.Inspector) error {
			return runTasks(reg, inspector, args[1:])
		})
	case "flush":
		return withInspector(cfg, func(inspector *asynq.Inspector) error {
			return runFlush(registry.ConsumeQueues(reg.AllSpecs()), inspector, args[1:])
		})
	case "routes":
		return runRoutes(cfg)
	case "help", "-h", "--help":
		printUsage(reg)
		return nil
	default:
		return runTaskCommand(cfg, reg, args[0], args[1:])
	}
}

func withInspector(cfg *config.Config, fn func(*asynq.Inspector) error) error {
	conn, err := appredis.AsynqConnOpt(cfg)
	if err != nil {
		return err
	}
	inspector := asynq.NewInspector(conn)
	err = fn(inspector)
	closeErr := inspector.Close()
	if err != nil {
		return err
	}
	return closeErr
}
