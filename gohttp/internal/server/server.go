package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"zengine/gohttp/internal/config"
	"zengine/gohttp/internal/server/devhandler"
	"zengine/gohttp/internal/server/prodhandler"
	"zengine/gohttp/internal/server/util"
)

const listenAddr = ":8080"

func New(cfg config.Config) (*http.Server, error) {
	log.Printf("gohttp start mode=%s listen=%s", cfg.Mode, listenAddr)

	mux := http.NewServeMux()
	injector := util.NewInjector(cfg.ZConfStatic)

	if cfg.Mode == "dev" {
		mux.Handle("/", devhandler.NewHandler(injector))
	} else {
		handler, err := prodhandler.NewHandler(injector)
		if err != nil {
			return nil, fmt.Errorf("build prod handler: %w", err)
		}
		mux.Handle("/", handler)
	}

		return &http.Server{
			Addr:              listenAddr,
			Handler:           loggingMiddleware(mux),
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
		}, nil
}
