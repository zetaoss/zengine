package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"
	appredis "github.com/zetaoss/zengine/goapp/app/redis"
	"github.com/zetaoss/zengine/goapp/server/runtime"
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
	redisClient, err := appredis.Open(cfg)
	if err != nil {
		return nil, err
	}
	serverCtx.JobEnqueuer = queue.New(redisClient)

	handler, err := BuildHandler(serverCtx)
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:              listenAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}, nil
}

func BuildHandler(serverCtx *serverctx.Context) (http.Handler, error) {
	components, err := runtime.NewComponents(serverCtx.Cfg)
	if err != nil {
		return nil, fmt.Errorf("build runtime components: %w", err)
	}

	mux := http.NewServeMux()
	if _, err := RegisterRoutes(mux, serverCtx, components); err != nil {
		return nil, fmt.Errorf("register routes: %w", err)
	}

	var handler http.Handler = mux
	handler = components.Wrap(handler)

	return handler, nil
}
