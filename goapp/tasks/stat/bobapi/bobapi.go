// Package bobapi calls bob, the in-cluster app server, for the stat tasks.
package bobapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var httpClient = &http.Client{Timeout: 60 * time.Second}

// Get requests bobEndpoint+path with the query params and decodes the "result" of an
// {"status":"ok","result":...} response into out.
func Get(ctx context.Context, bobEndpoint, path string, params url.Values, out any) error {
	if bobEndpoint == "" {
		return fmt.Errorf("BOB_ENDPOINT is required")
	}
	u := bobEndpoint + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	var body struct {
		Status string          `json:"status"`
		Error  string          `json:"error"`
		Result json.RawMessage `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("bob %s: status %d: %w", path, resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK || body.Status != "ok" {
		return fmt.Errorf("bob %s: status %d: %s", path, resp.StatusCode, body.Error)
	}
	return json.Unmarshal(body.Result, out)
}
