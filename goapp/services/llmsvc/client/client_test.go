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
		_, _ = w.Write([]byte(`{"model": "test-model", "choices":[{"message":{"content":"hello world"}}]}`))
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
