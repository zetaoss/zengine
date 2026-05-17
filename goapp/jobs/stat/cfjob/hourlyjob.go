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

type HourlyJob struct{}

const (
	hourlyJobName    = "stat-cf-hourly"
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

	token := jobCtx.Config().Cloudflare.APIToken
	zoneID := jobCtx.Config().Cloudflare.ZoneID
	if token == "" || zoneID == "" {
		return job.Error(fmt.Errorf("missing cloudflare credentials"))
	}

	until := timeutil.HourlyEndUTC(time.Now().UTC(), 10).Add(time.Hour)
	since := until.Add(-48 * time.Hour)

	rowsByTimeslot := map[string]map[string]string{}
	for cursor := since; cursor.Before(until); cursor = cursor.Add(24 * time.Hour) {
		winStart := cursor
		winEnd := cursor.Add(24 * time.Hour)
		if winEnd.After(until) {
			winEnd = until
		}

		payload, err := RunCFGraphQL(ctx, token, CFHourlyQuery, app.H{
			"zoneTag": zoneID,
			"since":   winStart.UTC().Format(time.RFC3339),
			"until":   winEnd.UTC().Format(time.RFC3339),
		})
		if err != nil {
			return job.Error(err)
		}

		for _, group := range CFGroups(payload) {
			timeslotRaw, _ := NestedString(group, "dimensions", "timeslot")
			if timeslotRaw == "" {
				continue
			}
			timeParsed, err := time.Parse(time.RFC3339, timeslotRaw)
			if err != nil {
				continue
			}
			timeslot := timeParsed.UTC().Format("2006-01-02 15:04:05")
			rowsByTimeslot[timeslot] = CFMetricsFromGroup(group)
		}
	}

	rows := make([]models.CFKV, 0, len(rowsByTimeslot)*len(models.WorkerCFMetricNames))
	for timeslot, metrics := range rowsByTimeslot {
		for _, name := range models.WorkerCFMetricNames {
			rows = append(rows, models.CFKV{Timeslot: timeslot, Name: name, Value: metrics[name]})
		}
	}

	if len(rows) > 0 {
		if err := db.Table("stat_cf_hourly").Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "timeslot"}, {Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).Create(&rows).Error; err != nil {
			return job.Error(err)
		}
	}

	return job.Success(app.H{"rows": len(rows)})
}

func (j *HourlyJob) Timeout() time.Duration { return hourlyJobTimeout }
