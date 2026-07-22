package gsc

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
	"github.com/zetaoss/zengine/goapp/models/stat"
	"github.com/zetaoss/zengine/goapp/tasks/stat/timeutil"

	"gorm.io/gorm/clause"
)

type HourlyTask struct{}

func NewHourlyTask() *HourlyTask {
	return &HourlyTask{}
}

func (j *HourlyTask) Execute(ctx context.Context, taskCtx taskctx.Context, _ any) (app.H, error) {
	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, err
	}
	sa, err := LoadGAReaderFile(taskCtx)
	if err != nil {
		return nil, err
	}
	siteURL := taskCtx.Config().Analytics.GSCSiteURL
	if siteURL == "" {
		siteURL = firstNonEmpty(sa.GscSiteURL, sa.GscSiteURL2, sa.SiteURL, sa.SiteURL2)
	}
	if siteURL == "" {
		return nil, fmt.Errorf("missing GSC site url")
	}
	slog.Debug("gsc hourly site url", "url", siteURL)

	token, err := FetchGoogleAccessToken(ctx, sa, "https://www.googleapis.com/auth/webmasters.readonly")
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("America/Los_Angeles")
	until := timeutil.HourlyEndInLocation(time.Now(), loc).Add(time.Hour)
	since := until.Add(-48 * time.Hour)

	body := app.H{
		"startDate":  since.Format("2006-01-02"),
		"endDate":    until.Add(-time.Hour).Format("2006-01-02"),
		"dimensions": []string{"hour"},
		"dataState":  "hourly_all",
		"rowLimit":   25000,
	}
	payload, err := RunGSCQuery(ctx, token, siteURL, body)
	if err != nil {
		return nil, err
	}
	rowsRaw, _ := payload["rows"].([]any)
	rows := make([]statmodels.GSC, 0, len(rowsRaw))
	for _, item := range rowsRaw {
		m, ok := item.(app.H)
		if !ok {
			continue
		}
		keys, _ := m["keys"].([]any)
		if len(keys) < 1 {
			continue
		}
		dateHourRaw, _ := keys[0].(string)
		tm, ok := parseGSCHour(dateHourRaw, loc)
		if !ok {
			continue
		}
		if tm.Before(since) || !tm.Before(until) {
			continue
		}
		rows = append(rows, statmodels.GSC{
			Timeslot:    tm.UTC().Format("2006-01-02 15:04:05"),
			Clicks:      asInt(m["clicks"]),
			Impressions: asInt(m["impressions"]),
			Ctr:         round4(asFloat(m["ctr"]) * 100),
			Position:    round4(asFloat(m["position"])),
		})
	}

	if len(rows) > 0 {
		if err := db.Table("stat_gsc_hourly").Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "timeslot"}}, DoUpdates: clause.AssignmentColumns([]string{"clicks", "impressions", "ctr", "position"})}).Create(&rows).Error; err != nil {
			return nil, err
		}
	}
	return app.H{"rows": len(rows)}, nil
}

func parseGSCHour(raw string, loc *time.Location) (time.Time, bool) {
	layouts := []string{time.RFC3339, "2006-01-02 15", "2006-01-02T15", "2006-01-02T15:04:05"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, raw, loc); err == nil {
			return t.In(loc).Truncate(time.Hour), true
		}
	}
	if len(raw) == 13 && strings.Contains(raw, " ") {
		if t, err := time.ParseInLocation("2006-01-02 15", raw, loc); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
