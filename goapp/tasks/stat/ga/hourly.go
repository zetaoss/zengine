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

type HourlyTask struct{}

func NewHourlyTask() *HourlyTask {
	return &HourlyTask{}
}

func (j *HourlyTask) Execute(ctx context.Context, taskCtx taskctx.Context, _ any) (app.H, error) {
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

	until := timeutil.HourlyEndInLocation(time.Now(), loc).Add(time.Hour)
	since := until.Add(-48 * time.Hour)

	token, err := gsc.FetchGoogleAccessToken(ctx, sa, "https://www.googleapis.com/auth/analytics.readonly")
	if err != nil {
		return nil, err
	}

	payload, err := runGAQuery(ctx, token, propertyID, since.Format("2006-01-02"), until.Format("2006-01-02"), []string{"date", "hour"})
	if err != nil {
		return nil, err
	}

	rows := parseGARows(payload, "20060102 15", "2006-01-02 15:04:05", loc, true)

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
