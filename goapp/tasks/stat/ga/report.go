package ga

import (
	"context"
	"net/url"
	"time"

	"github.com/zetaoss/zengine/goapp/models/stat"
	"github.com/zetaoss/zengine/goapp/tasks/stat/bobapi"
)

// report asks bob for GA4 rows. interval is "hour" (timeslot RFC3339 UTC) or "day" (timeslot
// YYYY-MM-DD); since and until are dates, both inclusive.
func report(ctx context.Context, bobEndpoint, interval string, since, until time.Time) ([]statmodels.GA, error) {
	var rows []struct {
		Timeslot        string `json:"timeslot"`
		Sessions        int    `json:"sessions"`
		ScreenPageViews int    `json:"screen_page_views"`
		ActiveUsers     int    `json:"active_users"`
	}
	params := url.Values{"interval": {interval}, "since": {since.Format("2006-01-02")}, "until": {until.Format("2006-01-02")}}
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

// location is the zone the GA windows are computed in (GA_TIMEZONE, default UTC).
func location(name string) *time.Location {
	if name == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}
