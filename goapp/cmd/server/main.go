package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("server config error", "err", err)
		os.Exit(1)
	}

	config.ConfigureSlog(cfg.App.LogLevel)
	slog.Debug("[server-main]", "cfg", cfg)

	srv, err := server.New(cfg)
	if err != nil {
		slog.Error("server build error", "err", err)
		os.Exit(1)
	}

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
