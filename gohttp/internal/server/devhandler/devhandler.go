package devhandler

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strings"

	"zengine/gohttp/internal/server/util"
)

func NewHandler(injector *util.Injector) http.Handler {
	target := url.URL{Scheme: "http", Host: "127.0.0.1:5173"}
	rp := httputil.NewSingleHostReverseProxy(&target)

	originalDirector := rp.Director
	rp.Director = func(req *http.Request) {
		log.Printf("dev request path=%s\nheaders:\n%s", req.URL.Path, formatHeaders(req.Header))
		originalDirector(req)
	}

	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error path=%s err=%v", r.URL.Path, err)
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}

	rp.ModifyResponse = func(resp *http.Response) error {
		// Do not touch upgraded/non-HTML responses (e.g. Vite HMR websocket).
		if resp.StatusCode == http.StatusSwitchingProtocols {
			return nil
		}
		ct := resp.Header.Get("Content-Type")
		if mt, _, err := mime.ParseMediaType(ct); err != nil || mt != "text/html" {
			return nil
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if err := resp.Body.Close(); err != nil {
			log.Printf("close err: %v", err)
		}

		mutated := injector.InjectScript(body, resp.Request)
		resp.Body = io.NopCloser(bytes.NewReader(mutated))
		resp.ContentLength = int64(len(mutated))
		resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(mutated)))
		return nil
	}

	return rp
}

func formatHeaders(h http.Header) string {
	if len(h) == 0 {
		return ""
	}

	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("  %s: %q", k, strings.Join(h.Values(k), ", ")))
	}
	return strings.Join(parts, "\n")
}
