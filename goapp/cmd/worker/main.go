package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
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
	slog.Info("starting worker")
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	runErr := wkr.Run(ctx)
	closeErr := wkr.Close()
	if runErr != nil {
		slog.Error("worker stopped with error", "err", runErr)
		os.Exit(1)
	}
	if closeErr != nil {
		slog.Error("worker close error", "err", closeErr)
		os.Exit(1)
	}
	slog.Info("worker stopped")
}
