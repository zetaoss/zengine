package mwjob

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"
	"github.com/zetaoss/zengine/goapp/jobs/stat/timeutil"
	"github.com/zetaoss/zengine/goapp/models"

	"gorm.io/gorm/clause"
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

type HourlyJobInput struct {
	Timeslot string `json:"timeslot"`
}

func (a *HourlyJob) Decode(raw []byte) (any, error) {
	var input HourlyJobInput
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &input); err != nil {
			return nil, err
		}
	}
	return input, nil
}

func (j *HourlyJob) Run(ctx context.Context, jobCtx job.JobContext, input HourlyJobInput) job.Result {
	db, err := jobCtx.GetDB()
	if err != nil {
		return job.Error(err)
	}

	stats, err := FetchMWStats(ctx, jobCtx)
	if err != nil {
		return job.Error(err)
	}

	ts := input.Timeslot
	if ts == "" {
		ts = timeutil.HourlyEndUTC(time.Now().UTC(), 0).Format("2006-01-02 15:04:05")
	}

	row := models.MWHourly{
		Timeslot:    ts,
		Articles:    ToInt(stats["articles"]),
		Pages:       ToInt(stats["pages"]),
		Images:      ToInt(stats["images"]),
		Edits:       ToInt(stats["edits"]),
		Users:       ToInt(stats["users"]),
		Activeusers: ToInt(stats["activeusers"]),
		Admins:      ToInt(stats["admins"]),
	}

	if err := db.Table("stat_mw_hourly").Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&row).Error; err != nil {
		return job.Error(err)
	}

	return job.Success(app.H{"row": row})
}

func (j *HourlyJob) Timeout() time.Duration { return hourlyJobTimeout }
