package gscjob

import (
	"context"
	"fmt"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"
	"github.com/zetaoss/zengine/goapp/jobs/stat/timeutil"
	"github.com/zetaoss/zengine/goapp/models"

	"gorm.io/gorm/clause"
)

type DailyJob struct{}

const (
	dailyJobName    = "stat-gsc-daily"
	dailyJobTimeout = 5 * time.Minute
)

func NewDailyJob() *DailyJob {
	return &DailyJob{}
}

func (j *DailyJob) Name() string { return dailyJobName }

func (j *DailyJob) Run(ctx context.Context, jobCtx job.JobContext, _ any) job.Result {
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
	to := timeutil.DailyEndInLocation(time.Now(), loc)
	from := to.AddDate(0, 0, -9)

	data := app.H{"startDate": from.Format("2006-01-02"), "endDate": to.Format("2006-01-02"), "dimensions": []string{"date"}, "rowLimit": 25000}
	payload, err := RunGSCQuery(ctx, token, siteURL, data)
	if err != nil {
		return job.Error(err)
	}
	rowsRaw, _ := payload["rows"].([]any)
	rows := make([]models.GSC, 0, len(rowsRaw))
	for _, item := range rowsRaw {
		m, _ := item.(app.H)
		keys, _ := m["keys"].([]any)
		if len(keys) < 1 {
			continue
		}
		d, _ := keys[0].(string)
		if len(d) != 10 {
			continue
		}
		rows = append(rows, models.GSC{Timeslot: d, Clicks: asInt(m["clicks"]), Impressions: asInt(m["impressions"]), Ctr: round4(asFloat(m["ctr"]) * 100), Position: round4(asFloat(m["position"]))})
	}

	if len(rows) > 0 {
		if err := db.Table("stat_gsc_daily").Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "timeslot"}}, DoUpdates: clause.AssignmentColumns([]string{"clicks", "impressions", "ctr", "position"})}).Create(&rows).Error; err != nil {
			return job.Error(err)
		}
	}
	return job.Success(app.H{"rows": len(rows)})
}

func (j *DailyJob) Timeout() time.Duration { return dailyJobTimeout }
