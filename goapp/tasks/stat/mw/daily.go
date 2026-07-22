package mw

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

	stats, err := FetchMWStats(ctx, taskCtx)
	if err != nil {
		return nil, err
	}

	row := statmodels.MWDaily{
		Timeslot:    timeutil.DailyEndUTC(time.Now().UTC()).Format("2006-01-02"),
		Articles:    ToInt(stats["articles"]),
		Pages:       ToInt(stats["pages"]),
		Images:      ToInt(stats["images"]),
		Edits:       ToInt(stats["edits"]),
		Users:       ToInt(stats["users"]),
		Activeusers: ToInt(stats["activeusers"]),
		Admins:      ToInt(stats["admins"]),
	}

	if err := db.Table("stat_mw_daily").Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&row).Error; err != nil {
		return nil, err
	}

	return app.H{"row": row}, nil
}
