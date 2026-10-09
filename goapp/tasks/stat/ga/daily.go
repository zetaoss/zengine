package ga

import (
	"context"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"

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
	// The last 10 property-local dates, today included.
	now := time.Now()
	rows, err := report(ctx, taskCtx.Config().API.BobEndpoint, "day", now.AddDate(0, 0, -9), now)
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
