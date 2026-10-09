package ga

import (
	"context"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
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
	cfg := taskCtx.Config()
	to := timeutil.DailyEndInLocation(time.Now(), location(cfg.Analytics.GATimezone))
	from := to.AddDate(0, 0, -9)

	rows, err := report(ctx, cfg.API.BobEndpoint, "day", from, to)
	if err != nil {
		return nil, err
	}

	if len(rows) > 0 {
		if err := db.Table("stat_ga_daily").Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "timeslot"}},
			DoUpdates: clause.AssignmentColumns([]string{"sessions", "screen_page_views", "active_users"}),
		}).Create(&rows).Error; err != nil {
			return nil, err
		}
	}

	return app.H{"rows": len(rows)}, nil
}
