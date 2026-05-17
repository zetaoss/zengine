package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w, code: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(sw, r)

		slog.Info("http",
			"method", r.Method,
			"uri", r.URL.RequestURI(),
			"status", sw.code,
			"dur", time.Since(start).Truncate(time.Millisecond),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	code int
}

func (w *statusWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(p []byte) (int, error) {
	return w.ResponseWriter.Write(p)
}
