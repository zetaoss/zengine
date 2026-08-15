package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/app/database"

	"gorm.io/gorm"
)

type dailyMetric struct {
	name    string
	integer bool
}

type dailyTable struct {
	name        string
	metrics     []dailyMetric
	valueColumn string
}

type dailySample struct {
	timeslot time.Time
	values   map[string]float64
}

var dailyStatTables = []dailyTable{
	{name: "stat_ga_daily", metrics: []dailyMetric{{"active_users", true}, {"screen_page_views", true}, {"sessions", true}}},
	{name: "stat_gsc_daily", metrics: []dailyMetric{{"clicks", true}, {"impressions", true}, {"ctr", false}, {"position", false}}},
	{name: "stat_mw_daily", metrics: []dailyMetric{{"pages", true}, {"articles", true}, {"edits", true}, {"images", true}, {"users", true}, {"activeusers", true}, {"admins", true}, {"jobs", true}}},
	{name: "stat_k8s_daily", metrics: []dailyMetric{{"node_cpu_usage", false}, {"node_cpu_allocatable", false}, {"node_memory_usage", false}, {"node_memory_allocatable", false}, {"pod_cpu_usage", false}, {"pod_memory_usage", false}, {"pvc_storage_usage", false}, {"pvc_storage_capacity", false}, {"pod_count", true}, {"defender_fighting_ratio", false}, {"defender_max_level", false}}},
	// Cloudflare stores one metric per row. Only scalar numeric metrics can be
	// interpolated; JSON breakdown metrics are intentionally left untouched.
	{name: "stat_cf_daily", metrics: []dailyMetric{{"value", false}}, valueColumn: "value"},
}

func runStatInterpolateDaily(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("stat-interpolate-daily", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	dryRun := fs.Bool("dry-run", false, "show rows without inserting them")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("stat-interpolate-daily does not accept positional arguments")
	}

	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer func() { _ = sqlDB.Close() }()

	total := 0
	for _, table := range dailyStatTables {
		if !db.Migrator().HasTable(table.name) {
			continue
		}
		inserted, err := interpolateDailyTable(db, table, *dryRun)
		if err != nil {
			return err
		}
		total += inserted
		_, _ = fmt.Printf("%s: %d %s\n", table.name, inserted, interpolationResultLabel(*dryRun))
	}
	_, _ = fmt.Printf("total: %d %s\n", total, interpolationResultLabel(*dryRun))
	return nil
}

func interpolationResultLabel(dryRun bool) string {
	if dryRun {
		return "rows to insert"
	}
	return "rows inserted"
}

func interpolateDailyTable(db *gorm.DB, table dailyTable, dryRun bool) (int, error) {
	if table.valueColumn != "" {
		return interpolateDailyKVTable(db, table, dryRun)
	}

	samples, err := loadDailySamples(db, table.name, table.metrics, "")
	if err != nil {
		return 0, err
	}
	rows := interpolateDailySamples(samples, table.metrics)
	if len(rows) == 0 || dryRun {
		return len(rows), nil
	}
	return len(rows), db.Table(table.name).Create(&rows).Error
}

func interpolateDailyKVTable(db *gorm.DB, table dailyTable, dryRun bool) (int, error) {
	var names []string
	if err := db.Table(table.name).Distinct().Pluck("name", &names).Error; err != nil {
		return 0, err
	}

	rows := make([]map[string]any, 0)
	for _, name := range names {
		samples, err := loadDailySamples(db.Where("name = ?", name), table.name, table.metrics, table.valueColumn)
		if err != nil {
			return 0, err
		}
		for _, row := range interpolateDailySamples(samples, table.metrics) {
			row["name"] = name
			row[table.valueColumn] = formatInterpolatedNumber(row[table.valueColumn].(float64))
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 || dryRun {
		return len(rows), nil
	}
	return len(rows), db.Table(table.name).Create(&rows).Error
}

func loadDailySamples(db *gorm.DB, tableName string, metrics []dailyMetric, valueColumn string) ([]dailySample, error) {
	columns := []string{"timeslot"}
	for _, metric := range metrics {
		columns = append(columns, metric.name)
	}
	var rawRows []map[string]any
	if err := db.Table(tableName).Select(columns).Order("timeslot ASC").Find(&rawRows).Error; err != nil {
		return nil, err
	}

	samples := make([]dailySample, 0, len(rawRows))
	for _, raw := range rawRows {
		timeslot, ok := parseDailyTimeslot(raw["timeslot"])
		if !ok {
			continue
		}
		values := make(map[string]float64, len(metrics))
		valid := true
		for _, metric := range metrics {
			value, ok := parseNumber(raw[metric.name])
			if !ok {
				valid = false
				break
			}
			values[metric.name] = value
		}
		if valid {
			samples = append(samples, dailySample{timeslot: timeslot, values: values})
		}
	}
	return samples, nil
}

func interpolateDailySamples(samples []dailySample, metrics []dailyMetric) []map[string]any {
	if len(samples) < 2 {
		return nil
	}
	sort.Slice(samples, func(i, j int) bool { return samples[i].timeslot.Before(samples[j].timeslot) })
	rows := make([]map[string]any, 0)
	for i := 0; i+1 < len(samples); i++ {
		left, right := samples[i], samples[i+1]
		days := int(right.timeslot.Sub(left.timeslot).Hours() / 24)
		if days <= 1 {
			continue
		}
		for offset := 1; offset < days; offset++ {
			ratio := float64(offset) / float64(days)
			row := map[string]any{"timeslot": left.timeslot.AddDate(0, 0, offset).Format("2006-01-02")}
			for _, metric := range metrics {
				value := left.values[metric.name] + (right.values[metric.name]-left.values[metric.name])*ratio
				if metric.integer {
					value = math.Round(value)
				}
				row[metric.name] = value
			}
			rows = append(rows, row)
		}
	}
	return rows
}

func parseDailyTimeslot(value any) (time.Time, bool) {
	switch v := value.(type) {
	case time.Time:
		return time.Date(v.UTC().Year(), v.UTC().Month(), v.UTC().Day(), 0, 0, 0, 0, time.UTC), true
	case []byte:
		return parseDailyTimeslot(string(v))
	case string:
		for _, layout := range []string{"2006-01-02", "2006-01-02 15:04:05", time.RFC3339} {
			if t, err := time.Parse(layout, v); err == nil {
				return time.Date(t.UTC().Year(), t.UTC().Month(), t.UTC().Day(), 0, 0, 0, 0, time.UTC), true
			}
		}
		return time.Time{}, false
	default:
		return time.Time{}, false
	}
}

func parseNumber(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int64:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int16:
		return float64(v), true
	case int8:
		return float64(v), true
	case uint64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint8:
		return float64(v), true
	case []byte:
		return parseNumber(string(v))
	case string:
		n, err := strconv.ParseFloat(v, 64)
		return n, err == nil
	default:
		return 0, false
	}
}

func formatInterpolatedNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
