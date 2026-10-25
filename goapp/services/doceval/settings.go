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

// The interval bounds, default model and strategies are config's, which validates the GOAPP_CONFIG file with
// the same rules.
const (
	MinActiveInterval = config.MinActiveInterval
	MaxInterval       = config.MaxInterval
	DefaultModel      = config.DefaultModel
	StrategyRandom    = config.StrategyRandom
	StrategyOldest    = config.StrategyOldest
	StrategyNewest    = config.StrategyNewest
)

// DefaultStrategy is the fallback candidate selection strategy (random, non-duplicate).
const DefaultStrategy = StrategyRandom

// Settings are the run settings of the two loops. Their models are the chains' (services/chainsvc).
type Settings struct {
	Interval string `json:"interval"`
	// Strategy is the candidate page selection strategy: "random", "oldest", or "newest".
	Strategy string `json:"strategy"`
	// DefineInterval and DefineStrategy run doc-define-loop, which writes the definitions evaluations need.
	DefineInterval string `json:"define_interval"`
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
// every 30 minutes, both with the random strategy.
func builtinSettings() Settings {
	return Settings{
		Interval:       "1h",
		Strategy:       DefaultStrategy,
		DefineInterval: "30m",
		DefineStrategy: DefaultStrategy,
	}
}

// DefaultSettings returns the defaults: chains.docEval and chains.docDefine of the GOAPP_CONFIG file, the
// strategy falling back to chains.common, and each field then to builtinSettings when unset or invalid. A nil cfg gives
// builtinSettings.
func DefaultSettings(cfg *config.Config) Settings {
	s := builtinSettings()
	if cfg == nil {
		return s
	}
	eval := cfg.Chains.DocEval.Resolve(cfg.Chains.Common)
	define := cfg.Chains.DocDefine.Resolve(cfg.Chains.Common)
	s.Interval, s.Strategy = resolved(eval, s.Interval, s.Strategy)
	s.DefineInterval, s.DefineStrategy = resolved(define, s.DefineInterval, s.DefineStrategy)
	return s
}

func resolved(c config.ChainConfig, interval, strategy string) (string, string) {
	interval = normalizedInterval(c.Interval, interval)
	if c.Strategy != "" {
		strategy = NormalizeStrategy(c.Strategy)
	}
	return interval, strategy
}

// ParseInterval parses a loop interval; see config.ParseInterval.
func ParseInterval(input string) (time.Duration, string, error) {
	return config.ParseInterval(input)
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
