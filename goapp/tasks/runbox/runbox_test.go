package runbox

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPageError(t *testing.T) {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
	}))
	defer slow.Close()
	_, timeoutErr := (&http.Client{Timeout: 10 * time.Millisecond}).Get(slow.URL)
	tests := []struct {
		name   string
		err    error
		status int
		bob    string
		want   string
	}{
		{"client timeout", timeoutErr, 0, "", errTimedOut},
		{"deadline", context.DeadlineExceeded, 0, "", errTimedOut},
		{"connection", errors.New(`Post "http://bob.example.internal/runbox/lang": dial tcp: connection refused`), 0, "", errUnavailable},
		{"bob 500", errors.New("runbox: dial tcp docker:2376"), 500, "runbox: dial tcp docker:2376", errUnavailable},
		{"bob 400", errors.New("invalid language"), 400, "invalid language", "invalid language"},
		{"bob 400 empty", errors.New("runbox http status=400"), 400, "", errUnavailable},
	}
	for _, tt := range tests {
		if got := pageError(tt.err, tt.status, tt.bob); got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}
