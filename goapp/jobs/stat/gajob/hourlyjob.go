package gajob

import (
	"context"
	"fmt"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"
	"github.com/zetaoss/zengine/goapp/jobs/stat/gscjob"
	"github.com/zetaoss/zengine/goapp/jobs/stat/timeutil"

	"gorm.io/gorm/clause"
)

type HourlyJob struct{}

const (
	hourlyJobName    = "stat-ga-hourly"
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
	sa, err := gscjob.LoadGAReaderFile(jobCtx)
	if err != nil {
		return job.Error(err)
	}
	propertyID := jobCtx.Config().Analytics.GAPropertyID
	if propertyID == "" {
		if sa.PropertyID != "" {
			propertyID = sa.PropertyID
		} else {
			propertyID = sa.PropertyID2
		}
	}
	if propertyID == "" {
		return job.Error(fmt.Errorf("missing GA property id"))
	}
	gaTZ := jobCtx.Config().Analytics.GATimezone
	if gaTZ == "" {
		gaTZ = "UTC"
	}
	loc, err := time.LoadLocation(gaTZ)
	if err != nil {
		loc = time.UTC
	}

	until := timeutil.HourlyEndInLocation(time.Now(), loc).Add(time.Hour)
	since := until.Add(-48 * time.Hour)

	token, err := gscjob.FetchGoogleAccessToken(ctx, sa, "https://www.googleapis.com/auth/analytics.readonly")
	if err != nil {
		return job.Error(err)
	}

	payload, err := runGAQuery(ctx, token, propertyID, since.Format("2006-01-02"), until.Format("2006-01-02"), []string{"date", "hour"})
	if err != nil {
		return job.Error(err)
	}

	rows := parseGARows(payload, "20060102 15", "2006-01-02 15:04:05")
	if len(rows) > 0 {
		if err := db.Table("stat_ga_hourly").Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "timeslot"}},
			DoUpdates: clause.AssignmentColumns([]string{"sessions", "pageviews", "users", "active_users"}),
		}).Create(&rows).Error; err != nil {
			return job.Error(err)
		}
	}

	return job.Success(app.H{"rows": len(rows)})
}

func (j *HourlyJob) Timeout() time.Duration { return hourlyJobTimeout }
