package ga

import (
	"context"
	"fmt"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
	"github.com/zetaoss/zengine/goapp/tasks/stat/gsc"
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
	sa, err := gsc.LoadGAReaderFile(taskCtx)
	if err != nil {
		return nil, err
	}
	propertyID := taskCtx.Config().Analytics.GAPropertyID
	if propertyID == "" {
		if sa.PropertyID != "" {
			propertyID = sa.PropertyID
		} else {
			propertyID = sa.PropertyID2
		}
	}
	if propertyID == "" {
		return nil, fmt.Errorf("missing GA property id")
	}
	gaTZ := taskCtx.Config().Analytics.GATimezone
	if gaTZ == "" {
		gaTZ = "UTC"
	}
	loc, err := time.LoadLocation(gaTZ)
	if err != nil {
		loc = time.UTC
	}

	to := timeutil.DailyEndInLocation(time.Now(), loc)
	from := to.AddDate(0, 0, -9)

	token, err := gsc.FetchGoogleAccessToken(ctx, sa, "https://www.googleapis.com/auth/analytics.readonly")
	if err != nil {
		return nil, err
	}

	payload, err := runGAQuery(ctx, token, propertyID, from.Format("2006-01-02"), to.Format("2006-01-02"), []string{"date"})
	if err != nil {
		return nil, err
	}

	rows := parseGARows(payload, "20060102", "2006-01-02", loc, false)
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
