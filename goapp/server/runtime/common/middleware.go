package common

import (
	"log/slog"
	"net/http"
	"time"
)

func AccessLogMiddleware(next http.Handler, wrapWriter func(*StatusWriter) http.ResponseWriter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &StatusWriter{ResponseWriter: w, Code: http.StatusOK}
		finalWriter := wrapWriter(sw)

		start := time.Now()
		next.ServeHTTP(finalWriter, r)

		slog.Info("http",
			"method", r.Method,
			"uri", r.URL.RequestURI(),
			"status", sw.Code,
			"dur", time.Since(start).Truncate(time.Millisecond),
		)
	})
}

type StatusWriter struct {
	http.ResponseWriter
	Code int
}

func (w *StatusWriter) WriteHeader(code int) {
	w.Code = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *StatusWriter) Write(p []byte) (int, error) {
	return w.ResponseWriter.Write(p)
}

func (w *StatusWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *StatusWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
