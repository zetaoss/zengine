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

type DailyTask struct{}

func NewDailyTask() *DailyTask {
	return &DailyTask{}
}

func (j *DailyTask) Execute(ctx context.Context, taskCtx taskctx.Context, _ any) (app.H, error) {
	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, err
	}

	to := timeutil.DailyEndUTC(time.Now().UTC())
	since := to.AddDate(0, 0, -9)
	until := to.AddDate(0, 0, 1)

	groups, err := FetchAnalytics(ctx, taskCtx.Config().API.BobEndpoint, "day", since.Format("2006-01-02"), until.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}

	rows := make([]statmodels.CFKV, 0, 2048)
	for _, group := range groups {
		for _, name := range statmodels.WorkerCFMetricNames {
			rows = append(rows, statmodels.CFKV{Timeslot: group.Timeslot, Name: name, Value: group.Metrics[name]})
		}
	}

	if len(rows) > 0 {
		if err := db.Table("stat_cf_daily").Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "timeslot"}, {Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).Create(&rows).Error; err != nil {
			return nil, err
		}
	}

	return app.H{"rows": len(rows)}, nil
}
