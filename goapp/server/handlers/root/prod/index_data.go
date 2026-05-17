package prod

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/zetaoss/zengine/goapp/server/handlers/root/injector"
)

const distDir = "/app/svelte/dist"

type IndexInjector struct {
	index    []byte
	injector *injector.Injector
}

func newIndexInjector(injector *injector.Injector) (*IndexInjector, error) {
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
		slog.Error("index write err", "err", err)
	}
}
