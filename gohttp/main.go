package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"zengine/gohttp/internal/config"
	"zengine/gohttp/internal/server"
)

func main() {
	cfg, err := config.Load(os.Args[1:])
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("server build error: %v", err)
	}

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
