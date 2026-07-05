package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/app/database"
	"github.com/zetaoss/zengine/goapp/worker"
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

	if err := run(cfg, wkr, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(cfg *config.Config, wkr *worker.Worker, args []string) error {
	if len(args) == 0 {
		printUsage(wkr)
		return nil
	}

	switch args[0] {
	case "migrate":
		return database.RunMigrate(cfg)
	case "jobs":
		return runJobs(wkr, args[1:])
	case "flush":
		return runFlush(wkr, args[1:])
	case "routes":
		return runRoutes(cfg)
	case "help", "-h", "--help":
		printUsage(wkr)
		return nil
	default:
		return runJobCommand(wkr, args[0], args[1:])
	}
}
