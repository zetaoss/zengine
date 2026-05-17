package cfjob

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
	dailyJobName    = "stat-cf-daily"
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

	token := jobCtx.Config().Cloudflare.APIToken
	zoneID := jobCtx.Config().Cloudflare.ZoneID
	if token == "" || zoneID == "" {
		return job.Error(fmt.Errorf("missing cloudflare credentials"))
	}

	to := timeutil.DailyEndUTC(time.Now().UTC())
	since := to.AddDate(0, 0, -9)
	until := to.AddDate(0, 0, 1)

	payload, err := RunCFGraphQL(ctx, token, CFDailyQuery, app.H{
		"zoneTag": zoneID,
		"since":   since.Format("2006-01-02"),
		"until":   until.Format("2006-01-02"),
	})
	if err != nil {
		return job.Error(err)
	}

	rows := make([]models.CFKV, 0, 2048)
	for _, group := range CFGroups(payload) {
		timeslot, _ := NestedString(group, "dimensions", "timeslot")
		if timeslot == "" {
			continue
		}
		metrics := CFMetricsFromGroup(group)
		for _, name := range models.WorkerCFMetricNames {
			rows = append(rows, models.CFKV{Timeslot: timeslot, Name: name, Value: metrics[name]})
		}
	}

	if len(rows) > 0 {
		if err := db.Table("stat_cf_daily").Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "timeslot"}, {Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).Create(&rows).Error; err != nil {
			return job.Error(err)
		}
	}

	return job.Success(app.H{"rows": len(rows)})
}

func (j *DailyJob) Timeout() time.Duration { return dailyJobTimeout }
