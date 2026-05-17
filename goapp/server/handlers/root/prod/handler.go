package prod

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/zetaoss/zengine/goapp/server/handlers/root/injector"
)

func New(injector *injector.Injector) (http.Handler, error) {
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
			indexInjector.ServeInjectedIndex(w, r)
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
