package doceval

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const SettingKeyDocEval = "doc_eval_settings"

// MinActiveInterval is the minimum active interval to prevent runaway loops.
const MinActiveInterval = 1 * time.Second

// MaxInterval is the strict upper bound (< 100h).
const MaxInterval = 100 * time.Hour

// DefaultModel is the bob aigate model used until a sysop picks another.
const DefaultModel = "groq/openai/gpt-oss-120b"

const (
	StrategyRandom = "random"
	StrategyOldest = "oldest"
	StrategyNewest = "newest"
)

// DefaultStrategy is the fallback candidate selection strategy (random, non-duplicate).
const DefaultStrategy = StrategyRandom

type Settings struct {
	Interval string `json:"interval"`
	// Model is the bob aigate model (provider/model) used for rubric generation and judging.
	Model string `json:"model"`
	// Strategy is the candidate page selection strategy: "random", "oldest", or "newest".
	Strategy string `json:"strategy"`
}

// NormalizeStrategy validates and normalizes a candidate selection strategy.
func NormalizeStrategy(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case StrategyOldest:
		return StrategyOldest
	case StrategyNewest:
		return StrategyNewest
	default:
		return StrategyRandom
	}
}

// DefaultSettings returns evaluation defaults (1h interval, DefaultModel, random strategy).
func DefaultSettings() Settings {
	return Settings{
		Interval: "1h",
		Model:    DefaultModel,
		Strategy: DefaultStrategy,
	}
}

// ParseInterval validates and parses an interval string.
// If input is "0s", "0", "off", or empty: returns 0, "0s", nil (disabled).
// Otherwise it parses via time.ParseDuration and verifies:
// - duration >= MinActiveInterval (1s)
// - duration < MaxInterval (100h)
func ParseInterval(input string) (time.Duration, string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" || strings.EqualFold(trimmed, "off") || trimmed == "0" || trimmed == "0s" || trimmed == "0m" || trimmed == "0h" {
		return 0, "0s", nil
	}

	d, err := time.ParseDuration(trimmed)
	if err != nil {
		return 0, "", fmt.Errorf("invalid duration format %q: %w (e.g. 10m, 1h, 6h, 24h, or 0s to disable)", input, err)
	}

	if d <= 0 {
		return 0, "0s", nil
	}

	if d < MinActiveInterval {
		return 0, "", fmt.Errorf("active interval must be at least %s, got %s (use 0s to disable)", MinActiveInterval, d)
	}

	if d >= MaxInterval {
		return 0, "", fmt.Errorf("interval must be less than 100h, got %s", d)
	}

	return d, trimmed, nil
}

// Duration returns the parsed duration and whether evaluation is active (> 0).
func (s Settings) Duration() (time.Duration, bool) {
	d, _, err := ParseInterval(s.Interval)
	if err != nil || d <= 0 {
		return 0, false
	}
	return d, true
}

func GetSettings(db *gorm.DB) (Settings, error) {
	var row models.AppSetting
	if err := db.Table("app_settings").Where("setting_key = ?", SettingKeyDocEval).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DefaultSettings(), nil
		}
		return Settings{}, err
	}
	var s Settings
	if err := json.Unmarshal([]byte(row.ValueJSON), &s); err != nil {
		return DefaultSettings(), nil
	}
	if _, norm, err := ParseInterval(s.Interval); err == nil {
		s.Interval = norm
	} else {
		s.Interval = DefaultSettings().Interval
	}
	if strings.TrimSpace(s.Model) == "" {
		s.Model = DefaultModel
	}
	s.Strategy = NormalizeStrategy(s.Strategy)
	return s, nil
}

func SaveSettings(db *gorm.DB, s Settings) error {
	_, norm, err := ParseInterval(s.Interval)
	if err != nil {
		return err
	}
	s.Interval = norm
	s.Strategy = NormalizeStrategy(s.Strategy)
	s.Model = strings.TrimSpace(s.Model)
	if provider, model, ok := strings.Cut(s.Model, "/"); !ok || provider == "" || model == "" {
		return fmt.Errorf("model must be provider/model, got %q", s.Model)
	}

	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}
	setting := models.AppSetting{
		SettingKey: SettingKeyDocEval,
		ValueJSON:  string(raw),
	}
	return db.Table("app_settings").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "setting_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value_json"}),
	}).Create(&setting).Error
}

const SettingKeyDocEvalLastRun = "doc_eval_last_run"

// LastRun is the outcome of the latest doc-eval-loop run, shown next to the interval setting.
type LastRun struct {
	At        time.Time `json:"at"`
	OK        bool      `json:"ok"`
	PageID    int       `json:"page_id,omitempty"`
	PageTitle string    `json:"page_title,omitempty"`
	Model     string    `json:"model,omitempty"`
	EvalID    uint64    `json:"eval_id,omitempty"`
	Error     string    `json:"error,omitempty"`
}

func GetLastRun(db *gorm.DB) (*LastRun, error) {
	var row models.AppSetting
	if err := db.Table("app_settings").Where("setting_key = ?", SettingKeyDocEvalLastRun).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	var r LastRun
	if err := json.Unmarshal([]byte(row.ValueJSON), &r); err != nil {
		return nil, nil
	}
	return &r, nil
}

func SaveLastRun(db *gorm.DB, r LastRun) error {
	raw, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return db.Table("app_settings").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "setting_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value_json"}),
	}).Create(&models.AppSetting{SettingKey: SettingKeyDocEvalLastRun, ValueJSON: string(raw)}).Error
}
