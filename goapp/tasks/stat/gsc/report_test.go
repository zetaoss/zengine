package gsc

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestQuery(t *testing.T) {
	var gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		_, _ = io.WriteString(w, `{"status":"ok","result":[{"timeslot":"2026-10-08","clicks":12,"impressions":340,"ctr":3.5294,"position":7.1235}]}`)
	}))
	defer srv.Close()

	rows, err := query(context.Background(), srv.URL, "day", time.Date(2026, 9, 29, 0, 0, 0, 0, pacific), time.Date(2026, 10, 8, 0, 0, 0, 0, pacific))
	if err != nil {
		t.Fatal(err)
	}
	if gotURL != "/gsc/query?interval=day&since=2026-09-29&until=2026-10-08" {
		t.Errorf("request URL = %s", gotURL)
	}
	m := rows[0].model(rows[0].Timeslot)
	if len(rows) != 1 || m.Timeslot != "2026-10-08" || m.Clicks != 12 || m.Impressions != 340 || m.Ctr != 3.5294 || m.Position != 7.1235 {
		t.Errorf("rows=%+v model=%+v", rows, m)
	}
}
