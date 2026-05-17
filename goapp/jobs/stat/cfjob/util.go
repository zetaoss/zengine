package cfjob

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/zetaoss/zengine/goapp/app"
	"io"
	"net/http"
	"time"
)

func RunCFGraphQL(ctx context.Context, token string, query string, variables app.H) (app.H, error) {
	body, _ := json.Marshal(app.H{"query": query, "variables": variables})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.cloudflare.com/client/v4/graphql", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("cloudflare api failed: %d %s", resp.StatusCode, string(raw))
	}

	var payload app.H
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	if errs, ok := payload["errors"].([]any); ok && len(errs) > 0 {
		return nil, fmt.Errorf("cloudflare graphql returned errors")
	}
	return payload, nil
}

func CFGroups(payload app.H) []app.H {
	data, _ := payload["data"].(app.H)
	viewer, _ := data["viewer"].(app.H)
	zones, _ := viewer["zones"].([]any)
	if len(zones) == 0 {
		return nil
	}
	zone0, _ := zones[0].(app.H)
	groups, _ := zone0["zones"].([]any)
	out := make([]app.H, 0, len(groups))
	for _, g := range groups {
		m, _ := g.(app.H)
		if m != nil {
			out = append(out, m)
		}
	}
	return out
}

func NestedString(m app.H, keys ...string) (string, bool) {
	cur := any(m)
	for _, k := range keys {
		next, ok := cur.(app.H)
		if !ok {
			return "", false
		}
		cur = next[k]
	}
	s, ok := cur.(string)
	return s, ok
}

func CFMetricsFromGroup(group app.H) map[string]string {
	sum, _ := group["sum"].(app.H)
	uniq, _ := group["uniq"].(app.H)

	metrics := map[string]string{}
	metrics["uniq_uniques"] = toTextValue(uniq["uniques"])
	metrics["sum_requests"] = toTextValue(sum["requests"])
	metrics["sum_pageViews"] = toTextValue(sum["pageViews"])
	metrics["sum_bytes"] = toTextValue(sum["bytes"])
	metrics["sum_cachedBytes"] = toTextValue(sum["cachedBytes"])
	metrics["sum_cachedRequests"] = toTextValue(sum["cachedRequests"])
	metrics["sum_encryptedBytes"] = toTextValue(sum["encryptedBytes"])
	metrics["sum_encryptedRequests"] = toTextValue(sum["encryptedRequests"])
	metrics["sum_threats"] = toTextValue(sum["threats"])
	metrics["sum_browserMap"] = toTextValue(sum["browserMap"])
	metrics["sum_contentTypeMap"] = toTextValue(sum["contentTypeMap"])
	metrics["sum_clientSSLMap"] = toTextValue(sum["clientSSLProtocol"])
	metrics["sum_countryMap"] = toTextValue(sum["countryMap"])
	metrics["sum_ipClassMap"] = toTextValue(sum["ipClassMap"])
	metrics["sum_responseStatusMap"] = toTextValue(sum["edgeResponseStatus"])
	metrics["sum_threatPathingMap"] = toTextValue(sum["threatPathingName"])
	return metrics
}

func toTextValue(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case app.H, []any:
		b, err := json.Marshal(x)
		if err != nil {
			return "[]"
		}
		return string(b)
	default:
		return fmt.Sprintf("%v", v)
	}
}

const CFDailyQuery = `query GetZoneAnalytics($zoneTag: string, $since: string, $until: string) {
  viewer {
    zones(filter: { zoneTag: $zoneTag }) {
      zones: httpRequests1dGroups(orderBy: [date_ASC], limit: 10000, filter: { date_geq: $since, date_lt: $until }) {
        dimensions { timeslot: date }
        uniq { uniques }
        sum {
          browserMap { pageViews key: uaBrowserFamily }
          bytes cachedBytes cachedRequests encryptedBytes encryptedRequests pageViews requests threats
          contentTypeMap { bytes requests key: edgeResponseContentTypeName }
          clientSSLMap { requests key: clientSSLProtocol }
          countryMap { bytes requests threats key: clientCountryName }
          ipClassMap { requests key: ipType }
          responseStatusMap { requests key: edgeResponseStatus }
          threatPathingMap { requests key: threatPathingName }
        }
      }
    }
  }
}`

const CFHourlyQuery = `query GetZoneAnalytics($zoneTag: string, $since: string, $until: string) {
  viewer {
    zones(filter: { zoneTag: $zoneTag }) {
      zones: httpRequests1hGroups(orderBy: [datetime_ASC], limit: 10000, filter: { datetime_geq: $since, datetime_lt: $until }) {
        dimensions { timeslot: datetime }
        uniq { uniques }
        sum {
          browserMap { pageViews key: uaBrowserFamily }
          bytes cachedBytes cachedRequests encryptedBytes encryptedRequests pageViews requests threats
          contentTypeMap { bytes requests key: edgeResponseContentTypeName }
          clientSSLMap { requests key: clientSSLProtocol }
          countryMap { bytes requests threats key: clientCountryName }
          ipClassMap { requests key: ipType }
          responseStatusMap { requests key: edgeResponseStatus }
          threatPathingMap { requests key: threatPathingName }
        }
      }
    }
  }
}`
