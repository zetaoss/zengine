package doceval

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"
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
	// DefineInterval, DefineModel and DefineStrategy run doc-define-loop, which writes the definitions
	// evaluations need.
	DefineInterval string `json:"define_interval"`
	DefineModel    string `json:"define_model"`
	DefineStrategy string `json:"define_strategy"`
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

// builtinSettings are the defaults when the deployment sets none: evaluation every hour and a definition
// every 30 minutes, both with DefaultModel and the random strategy.
func builtinSettings() Settings {
	return Settings{
		Interval:       "1h",
		Model:          DefaultModel,
		Strategy:       DefaultStrategy,
		DefineInterval: "30m",
		DefineModel:    DefaultModel,
		DefineStrategy: DefaultStrategy,
	}
}

// DefaultSettings returns the defaults: chains.docEval and chains.docDefine of the GOAPP_CONFIG file, each
// field falling back to chains.common and then to builtinSettings when unset or invalid. A nil cfg gives
// builtinSettings.
func DefaultSettings(cfg *config.Config) Settings {
	s := builtinSettings()
	if cfg == nil {
		return s
	}
	eval := cfg.Chains.DocEval.Resolve(cfg.Chains.Common)
	define := cfg.Chains.DocDefine.Resolve(cfg.Chains.Common)
	s.Interval, s.Model, s.Strategy = resolved(eval, s.Interval, s.Model, s.Strategy)
	s.DefineInterval, s.DefineModel, s.DefineStrategy = resolved(define, s.DefineInterval, s.DefineModel, s.DefineStrategy)
	return s
}

func resolved(c config.ChainConfig, interval, model, strategy string) (string, string, string) {
	interval = normalizedInterval(c.Interval, interval)
	if m, err := validateModel(c.Model); err == nil {
		model = m
	}
	if c.Strategy != "" {
		strategy = NormalizeStrategy(c.Strategy)
	}
	return interval, model, strategy
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
	return activeDuration(s.Interval)
}

// DefineDuration returns the parsed definition interval and whether doc-define-loop is active (> 0).
func (s Settings) DefineDuration() (time.Duration, bool) {
	return activeDuration(s.DefineInterval)
}

func activeDuration(interval string) (time.Duration, bool) {
	d, _, err := ParseInterval(interval)
	if err != nil || d <= 0 {
		return 0, false
	}
	return d, true
}

// GetSettings returns the saved settings over DefaultSettings(cfg): a field missing from the saved JSON (such
// as one added later) keeps its default, and an invalid saved value falls back to it.
func GetSettings(db *gorm.DB, cfg *config.Config) (Settings, error) {
	defaults := DefaultSettings(cfg)
	var row models.AppSetting
	if err := db.Table("app_settings").Where("setting_key = ?", SettingKeyDocEval).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return defaults, nil
		}
		return Settings{}, err
	}
	s := defaults
	if err := json.Unmarshal([]byte(row.ValueJSON), &s); err != nil {
		return defaults, nil
	}
	s.Interval = normalizedInterval(s.Interval, defaults.Interval)
	s.DefineInterval = normalizedInterval(s.DefineInterval, defaults.DefineInterval)
	if m, err := validateModel(s.Model); err == nil {
		s.Model = m
	} else {
		s.Model = defaults.Model
	}
	if m, err := validateModel(s.DefineModel); err == nil {
		s.DefineModel = m
	} else {
		s.DefineModel = defaults.DefineModel
	}
	s.Strategy = NormalizeStrategy(s.Strategy)
	s.DefineStrategy = NormalizeStrategy(s.DefineStrategy)
	return s, nil
}

func normalizedInterval(interval, fallback string) string {
	if strings.TrimSpace(interval) == "" {
		return fallback
	}
	if _, norm, err := ParseInterval(interval); err == nil {
		return norm
	}
	return fallback
}

func validateModel(model string) (string, error) {
	model = strings.TrimSpace(model)
	if provider, name, ok := strings.Cut(model, "/"); !ok || provider == "" || name == "" {
		return "", fmt.Errorf("model must be provider/model, got %q", model)
	}
	return model, nil
}

func SaveSettings(db *gorm.DB, s Settings) error {
	var err error
	if _, s.Interval, err = ParseInterval(s.Interval); err != nil {
		return err
	}
	if _, s.DefineInterval, err = ParseInterval(s.DefineInterval); err != nil {
		return fmt.Errorf("define interval: %w", err)
	}
	s.Strategy = NormalizeStrategy(s.Strategy)
	s.DefineStrategy = NormalizeStrategy(s.DefineStrategy)
	if s.Model, err = validateModel(s.Model); err != nil {
		return err
	}
	if s.DefineModel, err = validateModel(s.DefineModel); err != nil {
		return fmt.Errorf("define model: %w", err)
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

const (
	SettingKeyDocEvalLastRun = "doc_eval_last_run"
	SettingKeyDefineLastRun  = "doc_define_last_run"
)

// LastRun is the outcome of the latest doc-eval-loop or doc-define-loop run, shown next to its settings.
type LastRun struct {
	At        time.Time `json:"at"`
	OK        bool      `json:"ok"`
	PageID    int       `json:"page_id,omitempty"`
	PageTitle string    `json:"page_title,omitempty"`
	Model     string    `json:"model,omitempty"`
	EvalID    uint64    `json:"eval_id,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// GetLastRun returns the last run stored under key (SettingKeyDocEvalLastRun or SettingKeyDefineLastRun).
func GetLastRun(db *gorm.DB, key string) (*LastRun, error) {
	var row models.AppSetting
	if err := db.Table("app_settings").Where("setting_key = ?", key).Take(&row).Error; err != nil {
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

func SaveLastRun(db *gorm.DB, key string, r LastRun) error {
	raw, err := json.Marshal(r)
	if err != nil {
		return err
	}
	return db.Table("app_settings").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "setting_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value_json"}),
	}).Create(&models.AppSetting{SettingKey: key, ValueJSON: string(raw)}).Error
}
