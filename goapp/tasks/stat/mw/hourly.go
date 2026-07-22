package mw

import (
	"context"
	"encoding/json"
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

type HourlyTaskPayload struct {
	Timeslot string `json:"timeslot"`
}

func (a *HourlyTask) Decode(raw []byte) (any, error) {
	var input HourlyTaskPayload
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &input); err != nil {
			return nil, err
		}
	}
	return input, nil
}

func (j *HourlyTask) Execute(ctx context.Context, taskCtx taskctx.Context, input HourlyTaskPayload) (app.H, error) {
	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, err
	}

	stats, err := FetchMWStats(ctx, taskCtx)
	if err != nil {
		return nil, err
	}

	ts := input.Timeslot
	if ts == "" {
		ts = timeutil.HourlyEndUTC(time.Now().UTC(), 0).Format("2006-01-02 15:04:05")
	}

	row := statmodels.MWHourly{
		Timeslot:    ts,
		Articles:    ToInt(stats["articles"]),
		Pages:       ToInt(stats["pages"]),
		Images:      ToInt(stats["images"]),
		Edits:       ToInt(stats["edits"]),
		Users:       ToInt(stats["users"]),
		Activeusers: ToInt(stats["activeusers"]),
		Admins:      ToInt(stats["admins"]),
	}

	if err := db.Table("stat_mw_hourly").Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&row).Error; err != nil {
		return nil, err
	}

	return app.H{"row": row}, nil
}
