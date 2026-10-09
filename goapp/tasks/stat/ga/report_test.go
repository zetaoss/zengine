package ga

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReport(t *testing.T) {
	var gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		_, _ = io.WriteString(w, `{"status":"ok","result":[{"timeslot":"2026-10-09T07:00:00Z","sessions":5,"screen_page_views":9,"active_users":3},{"timeslot":"bad","sessions":1}]}`)
	}))
	defer srv.Close()

	until := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	rows, err := report(context.Background(), srv.URL, "hour", until.Add(-48*time.Hour), until)
	if err != nil {
		t.Fatal(err)
	}
	if gotURL != "/ga/report?interval=hour&since=2026-10-07T12%3A00%3A00Z&until=2026-10-09T12%3A00%3A00Z" {
		t.Errorf("request URL = %s", gotURL)
	}
	if len(rows) != 1 || rows[0].Timeslot != "2026-10-09 07:00:00" || rows[0].Sessions != 5 || rows[0].ScreenPageViews != 9 || rows[0].ActiveUsers != 3 {
		t.Errorf("rows=%+v", rows)
	}
}
