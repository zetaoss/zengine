package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestChatCompletionSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model": "test-model", "choices":[{"message":{"content":"hello world"}}], "usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}}`))
	}))
	defer ts.Close()

	svc := NewClientWithEndpoint(ts.URL)
	out, model, err := svc.ChatCompletion(context.Background(), "", []Message{{Role: "user", Content: "test"}})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out != "hello world" {
		t.Fatalf("unexpected output: %q", out)
	}
	if model != "test-model" {
		t.Fatalf("unexpected model: %q", model)
	}

	full, err := svc.ChatCompletionFull(context.Background(), "", []Message{{Role: "user", Content: "test"}})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if full.Usage.PromptTokens != 10 || full.Usage.CompletionTokens != 5 || full.Usage.TotalTokens != 15 {
		t.Fatalf("unexpected usage: %+v", full.Usage)
	}
}

func TestChatCompletionHTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer ts.Close()

	svc := NewClientWithEndpoint(ts.URL)
	_, _, err := svc.ChatCompletion(context.Background(), "", []Message{{Role: "user", Content: "test"}})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" || r.URL.RawQuery != "all" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":"not found"}`))
			return
		}
		_, _ = w.Write([]byte(`{"mode":"available","data":[]}`))
	}))
	defer ts.Close()

	svc := NewClientWithEndpoint(ts.URL)
	body, status, err := svc.GetJSON(context.Background(), "/v1/models?all")
	if err != nil || status != http.StatusOK || string(body) != `{"mode":"available","data":[]}` {
		t.Fatalf("unexpected result: %s %d %v", body, status, err)
	}
	if _, status, err := svc.GetJSON(context.Background(), "/v1/usage"); err != nil || status != http.StatusNotFound {
		t.Fatalf("expected 404, got %d %v", status, err)
	}
}

func rateLimitedServer(t *testing.T, retryAfter string, failures int) (*httptest.Server, *int) {
	t.Helper()
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls <= failures {
			if retryAfter != "" {
				w.Header().Set("Retry-After", retryAfter)
			}
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"provider groq returned status 429: tokens per minute"}`))
			return
		}
		_, _ = w.Write([]byte(`{"model":"m","choices":[{"message":{"content":"ok"}}]}`))
	}))
	t.Cleanup(ts.Close)
	return ts, &calls
}

func TestChatCompletionRetriesRateLimitOnce(t *testing.T) {
	ts, calls := rateLimitedServer(t, "5", 1)
	svc := NewClientWithEndpoint(ts.URL)
	var waited time.Duration
	svc.wait = func(_ context.Context, d time.Duration) error { waited = d; return nil }

	resp, err := svc.ChatCompletionFull(context.Background(), "m", nil)
	if err != nil || resp.Content != "ok" || *calls != 2 || waited != 5*time.Second {
		t.Fatalf("expected one retry after Retry-After, got %v %+v calls=%d waited=%v", err, resp, *calls, waited)
	}
}

func TestChatCompletionRateLimitGivesUp(t *testing.T) {
	cases := []struct {
		name       string
		retryAfter string
		failures   int
		wantCalls  int
		wantWait   time.Duration
	}{
		{"still limited after the retry", "", 2, 2, defaultRateLimitWait},
		{"delay longer than the cap", "3600", 1, 1, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ts, calls := rateLimitedServer(t, c.retryAfter, c.failures)
			svc := NewClientWithEndpoint(ts.URL)
			var waited time.Duration
			svc.wait = func(_ context.Context, d time.Duration) error { waited = d; return nil }

			_, err := svc.ChatCompletionFull(context.Background(), "m", nil)
			var statusErr *StatusError
			if !errors.As(err, &statusErr) || statusErr.StatusCode != http.StatusTooManyRequests || *calls != c.wantCalls || waited != c.wantWait {
				t.Fatalf("got %v calls=%d waited=%v", err, *calls, waited)
			}
			if !strings.Contains(err.Error(), "status=429: provider groq returned status 429: tokens per minute") {
				t.Fatalf("expected aigate's error in the message, got %q", err)
			}
		})
	}
}

func TestChatCompletionRateLimitWaitCanceled(t *testing.T) {
	ts, calls := rateLimitedServer(t, "1", 1)
	svc := NewClientWithEndpoint(ts.URL)
	svc.wait = func(context.Context, time.Duration) error { return context.Canceled }
	if _, err := svc.ChatCompletionFull(context.Background(), "m", nil); err == nil || *calls != 1 {
		t.Fatalf("expected a canceled wait to fail without retrying, got %v calls=%d", err, *calls)
	}
}
