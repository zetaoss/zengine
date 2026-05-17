package mwjob

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"
)

type mwStatsPayload struct {
	Query struct {
		Statistics app.H `json:"statistics"`
	} `json:"query"`
}

func FetchMWStats(ctx context.Context, jobCtx job.JobContext) (app.H, error) {
	apiServer := jobCtx.Config().App.APIServer
	if apiServer == "" {
		return nil, fmt.Errorf("apiServer is required")
	}

	u, err := url.Parse(apiServer)
	if err != nil {
		return nil, err
	}
	u.Path = "/w/api.php"

	q := u.Query()
	q.Set("action", "query")
	q.Set("meta", "siteinfo")
	q.Set("siprop", "statistics")
	q.Set("format", "json")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mediawiki api failed: %d %s", resp.StatusCode, string(b))
	}

	var payload mwStatsPayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	if len(payload.Query.Statistics) == 0 {
		return nil, fmt.Errorf("mediawiki response missing query.statistics")
	}

	return payload.Query.Statistics, nil
}

func ToInt(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case float32:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	case string:
		var n int
		_, _ = fmt.Sscanf(x, "%d", &n)
		return n
	default:
		return 0
	}
}
