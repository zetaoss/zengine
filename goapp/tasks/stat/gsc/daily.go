package gsc

import (
	"context"
	"fmt"
	"log/slog"
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
	sa, err := LoadGAReaderFile(taskCtx)
	if err != nil {
		return nil, err
	}
	siteURL := taskCtx.Config().Analytics.GSCSiteURL
	if siteURL == "" {
		siteURL = firstNonEmpty(sa.GscSiteURL, sa.GscSiteURL2, sa.SiteURL, sa.SiteURL2)
	}
	if siteURL == "" {
		return nil, fmt.Errorf("missing GSC site url")
	}
	slog.Debug("gsc daily site url", "url", siteURL)

	token, err := FetchGoogleAccessToken(ctx, sa, "https://www.googleapis.com/auth/webmasters.readonly")
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("America/Los_Angeles")
	to := timeutil.DailyEndInLocation(time.Now(), loc)
	from := to.AddDate(0, 0, -9)

	data := app.H{"startDate": from.Format("2006-01-02"), "endDate": to.Format("2006-01-02"), "dimensions": []string{"date"}, "rowLimit": 25000}
	payload, err := RunGSCQuery(ctx, token, siteURL, data)
	if err != nil {
		return nil, err
	}
	rowsRaw, _ := payload["rows"].([]any)
	rows := make([]statmodels.GSC, 0, len(rowsRaw))
	for _, item := range rowsRaw {
		m, ok := item.(app.H)
		if !ok {
			continue
		}
		keys, _ := m["keys"].([]any)
		if len(keys) < 1 {
			continue
		}
		d, _ := keys[0].(string)
		if len(d) != 10 {
			continue
		}
		rows = append(rows, statmodels.GSC{
			Timeslot:    d,
			Clicks:      asInt(m["clicks"]),
			Impressions: asInt(m["impressions"]),
			Ctr:         round4(asFloat(m["ctr"]) * 100),
			Position:    round4(asFloat(m["position"])),
		})
	}

	if len(rows) > 0 {
		if err := db.Table("stat_gsc_daily").Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "timeslot"}}, DoUpdates: clause.AssignmentColumns([]string{"clicks", "impressions", "ctr", "position"})}).Create(&rows).Error; err != nil {
			return nil, err
		}
	}
	return app.H{"rows": len(rows)}, nil
}
