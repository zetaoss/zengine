package main

import (
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/zetaoss/zengine/goapp/app/config"
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

	slog.Info("starting worker", "jobs", strings.Join(wkr.JobNames(), ", "))
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("stopping worker")
		wkr.Stop()
	}()

	wkr.Run()
	slog.Info("worker stopped")
}
