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

type HourlyTask struct{}

func NewHourlyTask() *HourlyTask {
	return &HourlyTask{}
}

func (j *HourlyTask) Execute(ctx context.Context, taskCtx taskctx.Context, _ any) (app.H, error) {
	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, err
	}

	token := taskCtx.Config().Cloudflare.APIToken
	zoneID := taskCtx.Config().Cloudflare.ZoneID
	if token == "" || zoneID == "" {
		return nil, fmt.Errorf("missing cloudflare credentials")
	}

	until := timeutil.HourlyEndUTC(time.Now().UTC(), 10).Add(time.Hour)
	since := until.Add(-48 * time.Hour)

	rowsByTimeslot := map[string]map[string]string{}
	for cursor := since; cursor.Before(until); cursor = cursor.Add(24 * time.Hour) {
		winStart := cursor
		winEnd := cursor.Add(24 * time.Hour)
		if winEnd.After(until) {
			winEnd = until
		}

		payload, err := RunCFGraphQL(ctx, token, CFHourlyQuery, app.H{
			"zoneTag": zoneID,
			"since":   winStart.UTC().Format(time.RFC3339),
			"until":   winEnd.UTC().Format(time.RFC3339),
		})
		if err != nil {
			return nil, err
		}

		for _, group := range CFGroups(payload) {
			timeslotRaw, _ := NestedString(group, "dimensions", "timeslot")
			if timeslotRaw == "" {
				continue
			}
			timeParsed, err := time.Parse(time.RFC3339, timeslotRaw)
			if err != nil {
				continue
			}
			timeslot := timeParsed.UTC().Format("2006-01-02 15:04:05")
			rowsByTimeslot[timeslot] = CFMetricsFromGroup(group)
		}
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
