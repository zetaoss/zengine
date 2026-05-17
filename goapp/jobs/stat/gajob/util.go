package gajob

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/models"
)

func runGAQuery(ctx context.Context, token, propertyID, startDate, endDate string, dimensions []string) (app.H, error) {
	endpoint := "https://analyticsdata.googleapis.com/v1beta/properties/" + propertyID + ":runReport"
	dims := make([]map[string]string, 0, len(dimensions))
	for _, d := range dimensions {
		dims = append(dims, map[string]string{"name": d})
	}
	body := app.H{
		"dateRanges": []map[string]string{{"startDate": startDate, "endDate": endDate}},
		"dimensions": dims,
		"metrics": []map[string]string{
			{"name": "sessions"},
			{"name": "screenPageViews"},
			{"name": "totalUsers"},
			{"name": "activeUsers"},
		},
		"keepEmptyRows": true,
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("ga api failed: %d %s", resp.StatusCode, string(b))
	}
	var payload app.H
	if err := json.Unmarshal(b, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func parseGARows(payload app.H, layout, outLayout string) []models.GA {
	rowsRaw, _ := payload["rows"].([]any)
	rows := make([]models.GA, 0, len(rowsRaw))
	for _, item := range rowsRaw {
		m, _ := item.(app.H)
		dimValues, _ := m["dimensionValues"].([]any)
		metricValues, _ := m["metricValues"].([]any)
		if len(dimValues) < 1 || len(metricValues) < 4 {
			continue
		}

		timeslotRaw := ""
		for i, dv := range dimValues {
			d, _ := dv.(app.H)
			val, _ := d["value"].(string)
			if i > 0 {
				timeslotRaw += " "
			}
			timeslotRaw += val
		}

		t, err := time.Parse(layout, timeslotRaw)
		if err != nil {
			continue
		}

		rows = append(rows, models.GA{
			Timeslot:        t.Format(outLayout),
			Sessions:        asInt(metricValues[0].(app.H)["value"]),
			ScreenPageViews: asInt(metricValues[1].(app.H)["value"]),
			ActiveUsers:     asInt(metricValues[3].(app.H)["value"]),
		})
	}
	return rows
}

func asInt(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case string:
		var n int
		_, _ = fmt.Sscanf(x, "%d", &n)
		return n
	default:
		return 0
	}
}
