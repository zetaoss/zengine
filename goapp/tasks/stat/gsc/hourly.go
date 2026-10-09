package gsc

import (
	"context"
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
	until := timeutil.HourlyEndInLocation(time.Now(), pacific).Add(time.Hour)
	since := until.Add(-48 * time.Hour)

	got, err := query(ctx, taskCtx.Config().API.BobEndpoint, "hour", since, until.Add(-time.Hour))
	if err != nil {
		return nil, err
	}
	rows := make([]statmodels.GSC, 0, len(got))
	for _, r := range got {
		tm, err := time.Parse(time.RFC3339, r.Timeslot)
		if err != nil || tm.Before(since) || !tm.Before(until) {
			continue
		}
		rows = append(rows, r.model(tm.UTC().Format("2006-01-02 15:04:05")))
	}

	if len(rows) > 0 {
		if err := db.Table("stat_gsc_hourly").Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "timeslot"}}, DoUpdates: clause.AssignmentColumns([]string{"clicks", "impressions", "ctr", "position"})}).Create(&rows).Error; err != nil {
			return nil, err
		}
	}
	return app.H{"rows": len(rows)}, nil
}
