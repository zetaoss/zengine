package doceval

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zetaoss/zengine/goapp/app/config"
)

func TestFetchWikiPageContentSuccess(t *testing.T) {
	mockAPI := `{
		"batchcomplete": true,
		"query": {
			"pages": [{
				"pageid": 6276,
				"title": "파이썬",
				"revisions": [{
					"revid": 845848,
					"slots": {
						"main": {
							"content": "==개요==\n파이썬 설명"
						}
					}
				}]
			}]
		}
	}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("rvprop") != "ids|content" {
			t.Errorf("expected rvprop=ids|content, got %q", r.URL.Query().Get("rvprop"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(mockAPI))
	}))
	defer ts.Close()

	cfg := &config.Config{
		App: config.AppConfig{
			APIServer: ts.URL,
		},
	}

	content, revid, err := fetchWikiPageContent(context.Background(), cfg, 6276)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if revid != 845848 || content != "==개요==\n파이썬 설명" {
		t.Fatalf("unexpected result: revid=%d, content=%q", revid, content)
	}
}

func TestFetchWikiPageContentRedirect(t *testing.T) {
	mockAPI := `{
		"batchcomplete": true,
		"query": {
			"pages": [{
				"pageid": 100,
				"redirect": true,
				"revisions": [{
					"revid": 1000,
					"slots": {
						"main": {
							"content": "#넘겨주기 [[파이썬]]"
						}
					}
				}]
			}]
		}
	}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("rvprop") != "ids|content" {
			t.Errorf("expected rvprop=ids|content, got %q", r.URL.Query().Get("rvprop"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(mockAPI))
	}))
	defer ts.Close()

	cfg := &config.Config{
		App: config.AppConfig{
			APIServer: ts.URL,
		},
	}

	_, _, err := fetchWikiPageContent(context.Background(), cfg, 100)
	if err == nil || err.Error() != "page 100 is a redirect" {
		t.Fatalf("expected redirect error, got: %v", err)
	}
}

func TestFetchWikiPageContentMissing(t *testing.T) {
	mockAPI := `{
		"batchcomplete": true,
		"query": {
			"pages": [{
				"missing": true
			}]
		}
	}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("rvprop") != "ids|content" {
			t.Errorf("expected rvprop=ids|content, got %q", r.URL.Query().Get("rvprop"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(mockAPI))
	}))
	defer ts.Close()

	cfg := &config.Config{
		App: config.AppConfig{
			APIServer: ts.URL,
		},
	}

	_, _, err := fetchWikiPageContent(context.Background(), cfg, 999999)
	if err == nil {
		t.Fatalf("expected error for missing page")
	}
}

func TestNormalizeWikiTitle(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"   ", ""},
		{"python", "Python"},
		{"Python", "Python"},
		{"hello world", "Hello_world"},
		{"hello   world", "Hello_world"},
		{"  _hello   _  world_  ", "Hello_world"},
		{"파이썬   문서", "파이썬_문서"},
	}

	for _, c := range cases {
		got := normalizeWikiTitle(c.in)
		if got != c.want {
			t.Errorf("normalizeWikiTitle(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
