package prodhandler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"zengine/gohttp/internal/server/util"
)

const distDir = "/app/svelte/dist"

type IndexInjector struct {
	index    []byte
	injector *util.Injector
}

func newIndexInjector(injector *util.Injector) (*IndexInjector, error) {
	indexPath := filepath.Join(distDir, "index.html")
	index, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("read index at startup: %w", err)
	}
	if len(index) == 0 {
		return nil, errors.New("empty index payload")
	}
	return &IndexInjector{index: index, injector: injector}, nil
}

func (ii *IndexInjector) ServeInjectedIndex(w http.ResponseWriter, r *http.Request) {
	injected := ii.injector.InjectScript(ii.index, r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method == http.MethodHead {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(injected)))
		w.WriteHeader(http.StatusOK)
		return
	}
	if _, err := w.Write(injected); err != nil {
		log.Printf("write err: %v", err)
	}
}
