package gsc

import (
	"context"
	"net/url"
	"time"

	"github.com/zetaoss/zengine/goapp/models/stat"
	"github.com/zetaoss/zengine/goapp/tasks/stat/bobapi"
)

// pacific is the zone Search Console dates and hours are in.
var pacific, _ = time.LoadLocation("America/Los_Angeles")

type row struct {
	Timeslot    string  `json:"timeslot"`
	Clicks      int     `json:"clicks"`
	Impressions int     `json:"impressions"`
	CTR         float64 `json:"ctr"`
	Position    float64 `json:"position"`
}

// query asks bob for Search Console rows. interval is "hour" (timeslot RFC3339 UTC) or "day"
// (timeslot YYYY-MM-DD); since and until are dates, both inclusive.
func query(ctx context.Context, bobEndpoint, interval string, since, until time.Time) ([]row, error) {
	var rows []row
	params := url.Values{"interval": {interval}, "since": {since.Format("2006-01-02")}, "until": {until.Format("2006-01-02")}}
	err := bobapi.Get(ctx, bobEndpoint, "/gsc/query", params, &rows)
	return rows, err
}

func (r row) model(timeslot string) statmodels.GSC {
	return statmodels.GSC{Timeslot: timeslot, Clicks: r.Clicks, Impressions: r.Impressions, Ctr: r.CTR, Position: r.Position}
}
