package cf

import (
	"context"
	"fmt"
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

	token := taskCtx.Config().Cloudflare.APIToken
	zoneID := taskCtx.Config().Cloudflare.ZoneID
	if token == "" || zoneID == "" {
		return nil, fmt.Errorf("missing cloudflare credentials")
	}

	to := timeutil.DailyEndUTC(time.Now().UTC())
	since := to.AddDate(0, 0, -9)
	until := to.AddDate(0, 0, 1)

	payload, err := RunCFGraphQL(ctx, token, CFDailyQuery, app.H{
		"zoneTag": zoneID,
		"since":   since.Format("2006-01-02"),
		"until":   until.Format("2006-01-02"),
	})
	if err != nil {
		return nil, err
	}

	rows := make([]statmodels.CFKV, 0, 2048)
	for _, group := range CFGroups(payload) {
		timeslot, _ := NestedString(group, "dimensions", "timeslot")
		if timeslot == "" {
			continue
		}
		metrics := CFMetricsFromGroup(group)
		for _, name := range statmodels.WorkerCFMetricNames {
			rows = append(rows, statmodels.CFKV{Timeslot: timeslot, Name: name, Value: metrics[name]})
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
