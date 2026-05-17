package mwjob

import (
	"context"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"
	"github.com/zetaoss/zengine/goapp/jobs/stat/timeutil"
	"github.com/zetaoss/zengine/goapp/models"
)

type HourlyJob struct{}

const (
	hourlyJobName    = "stat-mw-hourly"
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

	stats, err := FetchMWStats(ctx, jobCtx)
	if err != nil {
		return job.Error(err)
	}

	row := models.MWHourly{
		Timeslot:    timeutil.HourlyEndUTC(time.Now().UTC(), 0).Format("2006-01-02 15:04:05"),
		Articles:    ToInt(stats["articles"]),
		Pages:       ToInt(stats["pages"]),
		Images:      ToInt(stats["images"]),
		Edits:       ToInt(stats["edits"]),
		Users:       ToInt(stats["users"]),
		Activeusers: ToInt(stats["activeusers"]),
		Admins:      ToInt(stats["admins"]),
	}

	if err := db.Table("stat_mw_hourly").Save(&row).Error; err != nil {
		return job.Error(err)
	}

	return job.Success(app.H{"row": row})
}

func (j *HourlyJob) Timeout() time.Duration { return hourlyJobTimeout }
