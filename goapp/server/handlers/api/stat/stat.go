package stat

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/server/serverctx"

	"gorm.io/gorm"
)

func parseDays(r *http.Request) (int, bool) {
	days, err := strconv.Atoi(r.PathValue("days"))
	if err != nil || (days != 15 && days != 90) {
		return 0, false
	}
	return days, true
}

func CFHourly(c *serverctx.Context) {
	to := hourlyEndUTC(time.Now().UTC(), 10)
	from := to.Add(-47 * time.Hour)
	rows, err := selectCFRows(c.DB, "stat_cf_hourly", from, to, false)
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(buildCFHourlyPayload(from, to, rows))
}

func CFDaily(c *serverctx.Context) {
	days, ok := parseDays(c.R)
	if !ok {
		http.NotFound(c.W, c.R)
		return
	}
	to := dailyEndUTC(time.Now().UTC())
	from := to.AddDate(0, 0, -(days - 1))
	rows, err := selectCFRows(c.DB, "stat_cf_daily", from, to, true)
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(buildCFDailyPayload(from, to, rows))
}

func GAHourly(c *serverctx.Context) {
	to := hourlyEndUTC(time.Now().UTC(), 10)
	from := to.Add(-47 * time.Hour)
	rows, err := selectGAMergedRows(c.DB, []string{"stat_hourly_ga", "stat_ga_hourly"}, from, to, false)
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(buildNumericHourlyPayload(from, to, rows, models.StatGAMetricNames, func(row models.StatNumeric, name string) *float64 {
		switch name {
		case "active_users":
			return row.ActiveUsers
		case "screen_page_views":
			return row.ScreenPageViews
		case "sessions":
			return row.Sessions
		default:
			return nil
		}
	}))
}

func GADaily(c *serverctx.Context) {
	days, ok := parseDays(c.R)
	if !ok {
		http.NotFound(c.W, c.R)
		return
	}
	gaTZ := c.Cfg.Analytics.GATimezone
	if gaTZ == "" {
		gaTZ = "UTC"
	}
	loc, err := time.LoadLocation(gaTZ)
	if err != nil {
		loc = time.UTC
	}
	to := dailyEndInLocation(time.Now(), loc)
	from := to.AddDate(0, 0, -(days - 1))
	rows, err := selectGAMergedRows(c.DB, []string{"stat_daily_ga", "stat_ga_daily"}, from, to, true)
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(buildNumericDailyPayload(from, to, rows, models.StatGAMetricNames, func(row models.StatNumeric, name string) *float64 {
		switch name {
		case "active_users":
			return row.ActiveUsers
		case "screen_page_views":
			return row.ScreenPageViews
		case "sessions":
			return row.Sessions
		default:
			return nil
		}
	}))
}

func GSCHourly(c *serverctx.Context) {
	to := hourlyEndUTC(time.Now().UTC(), 10)
	from := to.Add(-47 * time.Hour)
	rows := make([]models.StatNumeric, 0, 256)
	err := c.DB.Table("stat_gsc_hourly").
		Select("timeslot", "clicks", "impressions", "ctr", "position").
		Where("timeslot BETWEEN ? AND ?", from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05")).
		Order("timeslot").
		Find(&rows).Error
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(buildNumericHourlyPayload(from, to, rows, models.StatGSCMetricNames, func(row models.StatNumeric, name string) *float64 {
		switch name {
		case "clicks":
			return row.Clicks
		case "impressions":
			return row.Impressions
		case "ctr":
			return row.Ctr
		case "position":
			return row.Position
		default:
			return nil
		}
	}))
}

func GSCDaily(c *serverctx.Context) {
	days, ok := parseDays(c.R)
	if !ok {
		http.NotFound(c.W, c.R)
		return
	}
	gscLoc, _ := time.LoadLocation("America/Los_Angeles")
	to := dailyEndInLocation(time.Now(), gscLoc)
	from := to.AddDate(0, 0, -(days - 1))
	rows := make([]models.StatNumeric, 0, 256)
	err := c.DB.Table("stat_gsc_daily").
		Select("timeslot", "clicks", "impressions", "ctr", "position").
		Where("DATE(timeslot) >= ?", from.Format("2006-01-02")).
		Where("DATE(timeslot) <= ?", to.Format("2006-01-02")).
		Order("timeslot").
		Find(&rows).Error
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(buildNumericDailyPayload(from, to, rows, models.StatGSCMetricNames, func(row models.StatNumeric, name string) *float64 {
		switch name {
		case "clicks":
			return row.Clicks
		case "impressions":
			return row.Impressions
		case "ctr":
			return row.Ctr
		case "position":
			return row.Position
		default:
			return nil
		}
	}))
}

func MWHourly(c *serverctx.Context) {
	to := hourlyEndUTC(time.Now().UTC(), 10)
	from := to.Add(-47 * time.Hour)
	rows := make([]models.StatNumeric, 0, 256)
	err := c.DB.Table("stat_mw_hourly").
		Select("timeslot", "pages", "articles", "edits", "images", "users", "activeusers", "admins", "jobs").
		Where("timeslot BETWEEN ? AND ?", from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05")).
		Order("timeslot").
		Find(&rows).Error
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(buildNumericHourlyPayload(from, to, rows, models.StatMWMetricNames, func(row models.StatNumeric, name string) *float64 {
		switch name {
		case "pages":
			return row.Pages
		case "articles":
			return row.Articles
		case "edits":
			return row.Edits
		case "images":
			return row.Images
		case "users":
			return row.Users
		case "activeusers":
			return row.ActiveUsersMW
		case "admins":
			return row.Admins
		case "jobs":
			return row.Jobs
		default:
			return nil
		}
	}))
}

func MWDaily(c *serverctx.Context) {
	days, ok := parseDays(c.R)
	if !ok {
		http.NotFound(c.W, c.R)
		return
	}
	to := dailyEndUTC(time.Now().UTC())
	from := to.AddDate(0, 0, -(days - 1))
	rows := make([]models.StatNumeric, 0, 256)
	err := c.DB.Table("stat_mw_daily").
		Select("timeslot", "pages", "articles", "edits", "images", "users", "activeusers", "admins", "jobs").
		Where("DATE(timeslot) >= ?", from.Format("2006-01-02")).
		Where("DATE(timeslot) <= ?", to.Format("2006-01-02")).
		Order("timeslot").
		Find(&rows).Error
	if err != nil {
		http.Error(c.W, "internal server error", http.StatusInternalServerError)
		return
	}
	c.JSON(buildNumericDailyPayload(from, to, rows, models.StatMWMetricNames, func(row models.StatNumeric, name string) *float64 {
		switch name {
		case "pages":
			return row.Pages
		case "articles":
			return row.Articles
		case "edits":
			return row.Edits
		case "images":
			return row.Images
		case "users":
			return row.Users
		case "activeusers":
			return row.ActiveUsersMW
		case "admins":
			return row.Admins
		case "jobs":
			return row.Jobs
		default:
			return nil
		}
	}))
}

func selectCFRows(db *gorm.DB, table string, from time.Time, to time.Time, dateOnly bool) ([]models.StatCF, error) {
	rows := make([]models.StatCF, 0, 512)
	q := db.Table(table).Select("timeslot", "name", "value").Where("name IN ?", models.StatCFMetricNames)
	if dateOnly {
		q = q.Where("DATE(timeslot) >= ?", from.Format("2006-01-02")).Where("DATE(timeslot) <= ?", to.Format("2006-01-02"))
	} else {
		q = q.Where("timeslot BETWEEN ? AND ?", from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05"))
	}
	err := q.Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func selectGAMergedRows(db *gorm.DB, tables []string, from time.Time, to time.Time, dateOnly bool) ([]models.StatNumeric, error) {
	merged := map[string]models.StatNumeric{}
	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			continue
		}
		rows := make([]models.StatNumeric, 0, 256)
		q := db.Table(table).
			Select("timeslot", "active_users", "screen_page_views", "sessions").
			Order("timeslot")
		if dateOnly {
			q = q.Where("DATE(timeslot) >= ?", from.Format("2006-01-02")).Where("DATE(timeslot) <= ?", to.Format("2006-01-02"))
		} else {
			q = q.Where("timeslot BETWEEN ? AND ?", from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05"))
		}
		if err := q.Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			merged[row.Timeslot.Format("2006-01-02 15:04:05")] = row
		}
	}
	out := make([]models.StatNumeric, 0, len(merged))
	for _, row := range merged {
		out = append(out, row)
	}
	return out, nil
}

func buildCFHourlyPayload(from time.Time, to time.Time, rows []models.StatCF) app.H {
	timeslots := make([]string, 0, 48)
	indexByTimeslot := map[string]int{}
	for cursor := from; !cursor.After(to); cursor = cursor.Add(time.Hour) {
		k := cursor.UTC().Format(time.RFC3339)
		indexByTimeslot[k] = len(timeslots)
		timeslots = append(timeslots, k)
	}
	series := emptySeriesByNames(models.StatCFMetricNames, len(timeslots))
	for _, row := range rows {
		timeslot := row.Timeslot.UTC().Format(time.RFC3339)
		idx, ok := indexByTimeslot[timeslot]
		if !ok {
			continue
		}
		values, ok := series[row.Name]
		if !ok {
			continue
		}
		values[idx] = parseJSONValue(row.Value)
		series[row.Name] = values
	}
	return mergePayload(timeslots, series, models.StatCFMetricNames)
}

func buildCFDailyPayload(from time.Time, to time.Time, rows []models.StatCF) app.H {
	timeslots := make([]string, 0, 90)
	indexByDate := map[string]int{}
	for cursor := from; !cursor.After(to); cursor = cursor.AddDate(0, 0, 1) {
		k := cursor.Format("2006-01-02")
		indexByDate[k] = len(timeslots)
		timeslots = append(timeslots, k)
	}
	series := emptySeriesByNames(models.StatCFMetricNames, len(timeslots))
	for _, row := range rows {
		dateKey := row.Timeslot.Format("2006-01-02")
		idx, ok := indexByDate[dateKey]
		if !ok {
			continue
		}
		values, ok := series[row.Name]
		if !ok {
			continue
		}
		values[idx] = parseJSONValue(row.Value)
		series[row.Name] = values
	}
	return mergePayload(timeslots, series, models.StatCFMetricNames)
}

func buildNumericHourlyPayload(from time.Time, to time.Time, rows []models.StatNumeric, names []string, pick func(models.StatNumeric, string) *float64) app.H {
	timeslots := make([]string, 0, 48)
	indexByTimeslot := map[string]int{}
	for cursor := from; !cursor.After(to); cursor = cursor.Add(time.Hour) {
		k := cursor.UTC().Format(time.RFC3339)
		indexByTimeslot[k] = len(timeslots)
		timeslots = append(timeslots, k)
	}
	series := emptySeriesByNames(names, len(timeslots))
	for _, row := range rows {
		timeslot := row.Timeslot.UTC().Format(time.RFC3339)
		idx, ok := indexByTimeslot[timeslot]
		if !ok {
			continue
		}
		for _, name := range names {
			if v := pick(row, name); v != nil {
				series[name][idx] = *v
			}
		}
	}
	return mergePayload(timeslots, series, names)
}

func buildNumericDailyPayload(from time.Time, to time.Time, rows []models.StatNumeric, names []string, pick func(models.StatNumeric, string) *float64) app.H {
	timeslots := make([]string, 0, 90)
	indexByDate := map[string]int{}
	for cursor := from; !cursor.After(to); cursor = cursor.AddDate(0, 0, 1) {
		k := cursor.Format("2006-01-02")
		indexByDate[k] = len(timeslots)
		timeslots = append(timeslots, k)
	}
	series := emptySeriesByNames(names, len(timeslots))
	for _, row := range rows {
		dateKey := row.Timeslot.Format("2006-01-02")
		idx, ok := indexByDate[dateKey]
		if !ok {
			continue
		}
		for _, name := range names {
			if v := pick(row, name); v != nil {
				series[name][idx] = *v
			}
		}
	}
	return mergePayload(timeslots, series, names)
}

func emptySeriesByNames(names []string, size int) map[string][]any {
	series := map[string][]any{}
	for _, name := range names {
		series[name] = make([]any, size)
	}
	return series
}

func mergePayload(timeslots []string, series map[string][]any, orderedNames []string) app.H {
	resp := app.H{"timeslots": timeslots}
	for _, name := range orderedNames {
		resp[name] = series[name]
	}
	return resp
}

func parseJSONValue(raw string) any {
	if raw == "" {
		return nil
	}
	var out any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}

func hourlyEndUTC(now time.Time, readyMinute int) time.Time {
	current := now.UTC()
	end := time.Date(current.Year(), current.Month(), current.Day(), current.Hour(), 0, 0, 0, time.UTC)
	if current.Minute() < readyMinute {
		end = end.Add(-time.Hour)
	}
	return end
}

func dailyEndUTC(now time.Time) time.Time {
	current := now.UTC()
	return time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, time.UTC)
}

func dailyEndInLocation(now time.Time, loc *time.Location) time.Time {
	current := now.In(loc)
	return time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, loc)
}
