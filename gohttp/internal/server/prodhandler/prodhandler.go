package prodhandler

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"zengine/gohttp/internal/server/util"
)

func NewHandler(injector *util.Injector) (http.Handler, error) {
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
		if cleanPath == "/" {
			indexInjector.ServeInjectedIndex(w, r)
			return
		}
		if !distPathExists(cleanPath) {
			indexInjector.ServeInjectedIndex(w, r)
			return
		}

		fs.ServeHTTP(w, r)
	}), nil
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
