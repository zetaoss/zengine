package mwjob

import (
	"context"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"
	"github.com/zetaoss/zengine/goapp/jobs/stat/timeutil"
	"github.com/zetaoss/zengine/goapp/models"
)

type DailyJob struct{}

const (
	dailyJobName    = "stat-mw-daily"
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

	stats, err := FetchMWStats(ctx, jobCtx)
	if err != nil {
		return job.Error(err)
	}

	row := models.MWDaily{
		Timeslot:    timeutil.DailyEndUTC(time.Now().UTC()).Format("2006-01-02"),
		Articles:    ToInt(stats["articles"]),
		Pages:       ToInt(stats["pages"]),
		Images:      ToInt(stats["images"]),
		Edits:       ToInt(stats["edits"]),
		Users:       ToInt(stats["users"]),
		Activeusers: ToInt(stats["activeusers"]),
		Admins:      ToInt(stats["admins"]),
	}

	if err := db.Table("stat_mw_daily").Save(&row).Error; err != nil {
		return job.Error(err)
	}

	return job.Success(app.H{"row": row})
}

func (j *DailyJob) Timeout() time.Duration { return dailyJobTimeout }
