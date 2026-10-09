package ga

import (
	"context"
	"net/url"
	"time"

	"github.com/zetaoss/zengine/goapp/models/stat"
	"github.com/zetaoss/zengine/goapp/tasks/stat/bobapi"
)

// report asks bob for GA4 rows in [since, until): the hours starting in it (timeslot RFC3339 UTC)
// or the property-local dates overlapping it (timeslot YYYY-MM-DD). bob converts with the GA
// property's time zone.
func report(ctx context.Context, bobEndpoint, interval string, since, until time.Time) ([]statmodels.GA, error) {
	var rows []struct {
		Timeslot        string `json:"timeslot"`
		Sessions        int    `json:"sessions"`
		ScreenPageViews int    `json:"screen_page_views"`
		ActiveUsers     int    `json:"active_users"`
	}
	params := url.Values{"interval": {interval}, "since": {since.UTC().Format(time.RFC3339)}, "until": {until.UTC().Format(time.RFC3339)}}
	if err := bobapi.Get(ctx, bobEndpoint, "/ga/report", params, &rows); err != nil {
		return nil, err
	}
	out := make([]statmodels.GA, 0, len(rows))
	for _, r := range rows {
		timeslot := r.Timeslot
		if interval == "hour" {
			t, err := time.Parse(time.RFC3339, r.Timeslot)
			if err != nil {
				continue
			}
			timeslot = t.UTC().Format("2006-01-02 15:04:05")
		}
		out = append(out, statmodels.GA{Timeslot: timeslot, Sessions: r.Sessions, ScreenPageViews: r.ScreenPageViews, ActiveUsers: r.ActiveUsers})
	}
	return out, nil
}
