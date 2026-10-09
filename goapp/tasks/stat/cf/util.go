package cf

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Group is one timeslot of Cloudflare zone analytics from bob: an RFC3339 datetime for hourly
// data, a date (YYYY-MM-DD) for daily data. Metric values are the text stored in stat_cf_*.
type Group struct {
	Timeslot string            `json:"timeslot"`
	Metrics  map[string]string `json:"metrics"`
}

var bobHTTPClient = &http.Client{Timeout: 60 * time.Second}

// FetchAnalytics asks bob for zone analytics. interval is "hour" (since/until RFC3339) or "day"
// (since/until YYYY-MM-DD, until exclusive).
func FetchAnalytics(ctx context.Context, bobEndpoint, interval, since, until string) ([]Group, error) {
	if bobEndpoint == "" {
		return nil, fmt.Errorf("BOB_ENDPOINT is required")
	}
	params := url.Values{"interval": {interval}, "since": {since}, "until": {until}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, bobEndpoint+"/cloudflare/analytics?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := bobHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var body struct {
		Status string  `json:"status"`
		Error  string  `json:"error"`
		Result []Group `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("bob cloudflare: status %d: %w", resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK || body.Status != "ok" {
		return nil, fmt.Errorf("bob cloudflare: status %d: %s", resp.StatusCode, body.Error)
	}
	return body.Result, nil
}
