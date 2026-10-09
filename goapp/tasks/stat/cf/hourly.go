package cf

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

	until := timeutil.HourlyEndUTC(time.Now().UTC(), 10).Add(time.Hour)
	since := until.Add(-48 * time.Hour)

	groups, err := FetchAnalytics(ctx, taskCtx.Config().API.BobEndpoint, "hour", since.Format(time.RFC3339), until.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}

	rowsByTimeslot := map[string]map[string]string{}
	for _, group := range groups {
		timeParsed, err := time.Parse(time.RFC3339, group.Timeslot)
		if err != nil {
			continue
		}
		rowsByTimeslot[timeParsed.UTC().Format("2006-01-02 15:04:05")] = group.Metrics
	}

	rows := make([]statmodels.CFKV, 0, len(rowsByTimeslot)*len(statmodels.WorkerCFMetricNames))
	for timeslot, metrics := range rowsByTimeslot {
		for _, name := range statmodels.WorkerCFMetricNames {
			rows = append(rows, statmodels.CFKV{Timeslot: timeslot, Name: name, Value: metrics[name]})
		}
	}

	if len(rows) > 0 {
		if err := db.Table("stat_cf_hourly").Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "timeslot"}, {Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).Create(&rows).Error; err != nil {
			return nil, err
		}
	}

	return app.H{"rows": len(rows)}, nil
}
