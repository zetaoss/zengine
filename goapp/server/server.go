package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/app/redis"
	"github.com/zetaoss/zengine/goapp/server/middleware"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
	"github.com/zetaoss/zengine/goapp/worker/queue"
)

const listenAddr = ":8080"

func New(cfg *config.Config) (*http.Server, error) {
	slog.Info("server start", "devMode", cfg.App.DevMode, "logLevel", cfg.App.LogLevel, "listen", listenAddr)

	serverCtx, err := serverctx.New(cfg)
	if err != nil {
		return nil, err
	}
	redisClient, err := redis.Open(cfg)
	if err != nil {
		return nil, err
	}
	enqueuer := queue.New(redisClient)
	serverCtx.JobEnqueuer = enqueuer

	mux := http.NewServeMux()
	if err := registerRoutes(mux, serverCtx); err != nil {
		return nil, fmt.Errorf("register routes: %w", err)
	}

	return &http.Server{
		Addr:              listenAddr,
		Handler:           middleware.Logging(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}, nil
}
