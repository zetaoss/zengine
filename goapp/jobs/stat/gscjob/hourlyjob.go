package gscjob

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"
	"github.com/zetaoss/zengine/goapp/jobs/stat/timeutil"
	"github.com/zetaoss/zengine/goapp/models"

	"gorm.io/gorm/clause"
)

type HourlyJob struct{}

const (
	hourlyJobName    = "stat-gsc-hourly"
	hourlyJobTimeout = 5 * time.Minute
)

func NewHourlyJob() *HourlyJob {
	return &HourlyJob{}
}

func (j *HourlyJob) Name() string { return hourlyJobName }

func (j *HourlyJob) Run(ctx context.Context, jobCtx job.JobContext, _ any) job.Result {
	db, err := jobCtx.GetDB()
	if err != nil {
		return job.Error(err)
	}
	sa, err := LoadGAReaderFile(jobCtx)
	if err != nil {
		return job.Error(err)
	}
	siteURL := jobCtx.Config().Analytics.GSCSiteURL
	if siteURL == "" {
		siteURL = firstNonEmpty(sa.GscSiteURL, sa.GscSiteURL2, sa.SiteURL, sa.SiteURL2)
	}
	if siteURL == "" {
		return job.Error(fmt.Errorf("missing GSC site url"))
	}

	token, err := FetchGoogleAccessToken(ctx, sa, "https://www.googleapis.com/auth/webmasters.readonly")
	if err != nil {
		return job.Error(err)
	}

	loc, _ := time.LoadLocation("America/Los_Angeles")
	until := timeutil.HourlyEndInLocation(time.Now(), loc).Add(time.Hour)
	since := until.Add(-48 * time.Hour)

	body := app.H{"startDate": since.Format("2006-01-02"), "endDate": until.Format("2006-01-02"), "dimensions": []string{"date", "hour"}, "rowLimit": 25000}
	payload, err := RunGSCQuery(ctx, token, siteURL, body)
	if err != nil {
		return job.Error(err)
	}
	rowsRaw, _ := payload["rows"].([]any)
	rows := make([]models.GSC, 0, len(rowsRaw))
	for _, item := range rowsRaw {
		m, _ := item.(app.H)
		keys, _ := m["keys"].([]any)
		if len(keys) < 2 {
			continue
		}
		dateRaw, _ := keys[0].(string)
		hourRaw, _ := keys[1].(string)
		raw := dateRaw + " " + hourRaw
		tm, ok := parseGSCHour(raw, loc)
		if !ok {
			continue
		}
		if tm.Before(since) || !tm.Before(until) {
			continue
		}
		rows = append(rows, models.GSC{Timeslot: tm.UTC().Format("2006-01-02 15:04:05"), Clicks: asInt(m["clicks"]), Impressions: asInt(m["impressions"]), Ctr: round4(asFloat(m["ctr"]) * 100), Position: round4(asFloat(m["position"]))})
	}

	if len(rows) > 0 {
		if err := db.Table("stat_gsc_hourly").Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "timeslot"}}, DoUpdates: clause.AssignmentColumns([]string{"clicks", "impressions", "ctr", "position"})}).Create(&rows).Error; err != nil {
			return job.Error(err)
		}
	}
	return job.Success(app.H{"rows": len(rows)})
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

func (j *HourlyJob) Timeout() time.Duration { return hourlyJobTimeout }
