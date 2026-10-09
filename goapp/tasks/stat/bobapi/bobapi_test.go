package bobapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	var gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		_, _ = io.WriteString(w, `{"status":"ok","result":[{"n":1}]}`)
	}))
	defer srv.Close()

	var out []struct{ N int }
	if err := Get(context.Background(), srv.URL, "/x/y", url.Values{"a": {"1"}}, &out); err != nil {
		t.Fatal(err)
	}
	if gotURL != "/x/y?a=1" || len(out) != 1 || out[0].N != 1 {
		t.Errorf("url=%s out=%+v", gotURL, out)
	}
}

func TestGetErrors(t *testing.T) {
	var out any
	if err := Get(context.Background(), "", "/x", nil, &out); err == nil || !strings.Contains(err.Error(), "BOB_ENDPOINT") {
		t.Errorf("empty endpoint: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, `{"status":"error","error":"upstream 403"}`)
	}))
	defer srv.Close()
	if err := Get(context.Background(), srv.URL, "/x", nil, &out); err == nil || !strings.Contains(err.Error(), "upstream 403") {
		t.Errorf("bob error: %v", err)
	}
}
