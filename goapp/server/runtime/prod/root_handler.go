package prod

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/zetaoss/zengine/goapp/server/runtime/common"
)

const distDir = "/app/svelte/dist"

func newRootHandler(injector *common.Injector) (http.Handler, error) {
	fs := http.FileServer(http.Dir(distDir))
	indexInjector, err := newIndexInjector(injector)
	if err != nil {
		return nil, err
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		cleanPath := path.Clean("/" + r.URL.Path)
		if shouldServeInjectedIndex(cleanPath, distPathExists(cleanPath)) {
			indexInjector.serveInjectedIndex(w, r)
			return
		}

		fs.ServeHTTP(w, r)
	}), nil
}

func shouldServeInjectedIndex(cleanPath string, exists bool) bool {
	return cleanPath == "/" || !exists
}

func distPathExists(cleanPath string) bool {
	localPath := filepath.Join(distDir, filepath.FromSlash(strings.TrimPrefix(cleanPath, "/")))
	_, err := os.Stat(localPath)
	if err == nil {
		return true
	}
	// Let FileServer handle non-not-found errors (e.g. permission issues).
	return !os.IsNotExist(err)
}

type indexInjector struct {
	index    []byte
	injector *common.Injector
}

func newIndexInjector(injector *common.Injector) (*indexInjector, error) {
	indexPath := filepath.Join(distDir, "index.html")
	index, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("read index at startup: %w", err)
	}
	if len(index) == 0 {
		return nil, errors.New("empty index payload")
	}
	return &indexInjector{index: index, injector: injector}, nil
}

func (ii *indexInjector) serveInjectedIndex(w http.ResponseWriter, r *http.Request) {
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
