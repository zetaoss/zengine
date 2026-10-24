package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
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
