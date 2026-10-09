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

type DailyTask struct{}

func NewDailyTask() *DailyTask {
	return &DailyTask{}
}

func (j *DailyTask) Execute(ctx context.Context, taskCtx taskctx.Context, _ any) (app.H, error) {
	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, err
	}
	to := timeutil.DailyEndInLocation(time.Now(), pacific)
	from := to.AddDate(0, 0, -9)

	got, err := query(ctx, taskCtx.Config().API.BobEndpoint, "day", from, to)
	if err != nil {
		return nil, err
	}
	rows := make([]statmodels.GSC, 0, len(got))
	for _, r := range got {
		rows = append(rows, r.model(r.Timeslot))
	}

	if len(rows) > 0 {
		if err := db.Table("stat_gsc_daily").Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "timeslot"}}, DoUpdates: clause.AssignmentColumns([]string{"clicks", "impressions", "ctr", "position"})}).Create(&rows).Error; err != nil {
			return nil, err
		}
	}
	return app.H{"rows": len(rows)}, nil
}
