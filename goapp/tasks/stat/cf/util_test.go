package cf

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchAnalytics(t *testing.T) {
	var gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		_, _ = io.WriteString(w, `{"status":"ok","result":[{"timeslot":"2026-10-09T07:00:00Z","metrics":{"uniq_uniques":"12","sum_countryMap":"[]"}}]}`)
	}))
	defer srv.Close()

	groups, err := FetchAnalytics(context.Background(), srv.URL, "hour", "2026-10-09T07:00:00Z", "2026-10-09T08:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	if gotURL != "/cloudflare/analytics?interval=hour&since=2026-10-09T07%3A00%3A00Z&until=2026-10-09T08%3A00%3A00Z" {
		t.Errorf("request URL = %s", gotURL)
	}
	if len(groups) != 1 || groups[0].Timeslot != "2026-10-09T07:00:00Z" || groups[0].Metrics["uniq_uniques"] != "12" {
		t.Errorf("groups=%+v", groups)
	}
}

func TestFetchAnalyticsErrors(t *testing.T) {
	if _, err := FetchAnalytics(context.Background(), "", "day", "a", "b"); err == nil || !strings.Contains(err.Error(), "BOB_ENDPOINT") {
		t.Errorf("empty endpoint: %v", err)
	}
}
