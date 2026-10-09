package ga

import (
	"context"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
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
	cfg := taskCtx.Config()
	loc := location(cfg.Analytics.GATimezone)
	until := timeutil.HourlyEndInLocation(time.Now(), loc).Add(time.Hour)
	since := until.Add(-48 * time.Hour)

	rows, err := report(ctx, cfg.API.BobEndpoint, "hour", since, until)
	if err != nil {
		return nil, err
	}

	if len(rows) > 0 {
		if err := db.Table("stat_ga_hourly").Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "timeslot"}},
			DoUpdates: clause.AssignmentColumns([]string{"sessions", "screen_page_views", "active_users"}),
		}).Create(&rows).Error; err != nil {
			return nil, err
		}
	}

	return app.H{"rows": len(rows)}, nil
}
