package prodhandler

import (
	"net/http"

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

		if r.URL.Path == "/" {
			indexInjector.ServeInjectedIndex(w, r)
			return
		}

		fs.ServeHTTP(w, r)
	}), nil
}
