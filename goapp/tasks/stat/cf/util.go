package cf

import (
	"context"
	"net/url"

	"github.com/zetaoss/zengine/goapp/tasks/stat/bobapi"
)

// Group is one timeslot of Cloudflare zone analytics from bob: an RFC3339 datetime for hourly
// data, a date (YYYY-MM-DD) for daily data. Metric values are the text stored in stat_cf_*.
type Group struct {
	Timeslot string            `json:"timeslot"`
	Metrics  map[string]string `json:"metrics"`
}

// FetchAnalytics asks bob for zone analytics. interval is "hour" (since/until RFC3339) or "day"
// (since/until YYYY-MM-DD, until exclusive).
func FetchAnalytics(ctx context.Context, bobEndpoint, interval, since, until string) ([]Group, error) {
	var groups []Group
	err := bobapi.Get(ctx, bobEndpoint, "/cloudflare/analytics", url.Values{"interval": {interval}, "since": {since}, "until": {until}}, &groups)
	return groups, err
}
