package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/worker/scheduler"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("scheduler config error", "err", err)
		os.Exit(1)
	}
	config.ConfigureSlog(cfg.App.LogLevel)

	runtime, err := scheduler.New(cfg)
	if err != nil {
		slog.Error("scheduler build error", "err", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slog.Info("starting scheduler")
	if err := runtime.Run(ctx); err != nil {
		slog.Error("scheduler stopped with error", "err", err)
		os.Exit(1)
	}
	slog.Info("scheduler stopped")
}
