package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	"github.com/zetaoss/zengine/goapp/app/config"
	appredis "github.com/zetaoss/zengine/goapp/app/redis"
	"github.com/zetaoss/zengine/goapp/server/runtime"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
)

const listenAddr = ":8080"

type Server struct {
	http       *http.Server
	taskClient *asynq.Client
	closeOnce  sync.Once
	closeErr   error
}

func New(cfg *config.Config) (*Server, error) {
	slog.Info("server start", "devMode", cfg.App.DevMode, "logLevel", cfg.App.LogLevel, "listen", listenAddr)

	serverCtx, err := serverctx.New(cfg)
	if err != nil {
		return nil, err
	}
	redisConn, err := appredis.AsynqConnOpt(cfg)
	if err != nil {
		return nil, err
	}
	taskClient := asynq.NewClient(redisConn)
	serverCtx.TaskEnqueuer = taskClient

	handler, err := BuildHandler(serverCtx)
	if err != nil {
		_ = taskClient.Close()
		return nil, err
	}

	return &Server{http: &http.Server{
		Addr:              listenAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}, taskClient: taskClient}, nil
}

func (s *Server) ListenAndServe() error { return s.http.ListenAndServe() }

func (s *Server) Shutdown(ctx context.Context) error {
	httpErr := s.http.Shutdown(ctx)
	closeErr := s.Close()
	if httpErr != nil {
		return httpErr
	}
	return closeErr
}

func (s *Server) Close() error {
	s.closeOnce.Do(func() { s.closeErr = s.taskClient.Close() })
	return s.closeErr
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
