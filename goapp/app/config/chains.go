package config

import (
	"fmt"
	"strings"
	"time"
)

// DefaultModel is the bob aigate model a chain runs on when neither its settings nor chains.common name one.
const DefaultModel = "groq/openai/gpt-oss-120b"

// Candidate strategies of the chain loops.
const (
	StrategyRandom = "random"
	StrategyOldest = "oldest"
	StrategyNewest = "newest"
)

const (
	// MinActiveInterval keeps a loop from running back to back.
	MinActiveInterval = 1 * time.Second
	// MaxInterval is the exclusive upper bound of a loop interval.
	MaxInterval = 100 * time.Hour
)

// ParseInterval parses a loop interval. "0s", "0", "off" or empty turn the loop off (0, "0s"); any other value
// is a Go duration from MinActiveInterval up to, not including, MaxInterval.
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

// ValidateModel returns model trimmed, or an error unless it is a bob aigate model ID, provider/model.
func ValidateModel(model string) (string, error) {
	model = strings.TrimSpace(model)
	if provider, name, ok := strings.Cut(model, "/"); !ok || provider == "" || name == "" {
		return "", fmt.Errorf("model must be provider/model, got %q", model)
	}
	return model, nil
}

// ValidStrategy reports whether s is a candidate strategy.
func ValidStrategy(s string) bool {
	switch s {
	case StrategyRandom, StrategyOldest, StrategyNewest:
		return true
	}
	return false
}

// ChainsConfig is the chains section of the GOAPP_CONFIG file: run settings of the chains. The loops (docDefine,
// docEval) take an interval, model and strategy; the on-demand chains (docCreate, docUpdate) take a model. A
// model or strategy left unset falls back to common. Settings a sysop saves in the evaluation tab override the
// loops' values.
type ChainsConfig struct {
	Common    CommonChainConfig `yaml:"common"`
	DocDefine ChainConfig       `yaml:"docDefine"`
	DocEval   ChainConfig       `yaml:"docEval"`
	DocCreate ModelChainConfig  `yaml:"docCreate"`
	DocUpdate ModelChainConfig  `yaml:"docUpdate"`
}

// CommonChainConfig holds what the chains share. It has no interval: each loop runs at its own pace.
type CommonChainConfig struct {
	Model    string `yaml:"model"`
	Strategy string `yaml:"strategy"`
}

// ChainConfig is the run settings of a chain loop.
type ChainConfig struct {
	Interval string `yaml:"interval"`
	Model    string `yaml:"model"`
	Strategy string `yaml:"strategy"`
}

// ModelChainConfig is the run settings of a chain run on request.
type ModelChainConfig struct {
	Model string `yaml:"model"`
}

// Resolve returns c with its unset model and strategy taken from common.
func (c ChainConfig) Resolve(common CommonChainConfig) ChainConfig {
	if c.Model == "" {
		c.Model = common.Model
	}
	if c.Strategy == "" {
		c.Strategy = common.Strategy
	}
	return c
}

// ModelFor returns the model of an on-demand chain: its own, else common's, else DefaultModel.
func (c ChainsConfig) ModelFor(chain ModelChainConfig) string {
	for _, m := range []string{chain.Model, c.Common.Model} {
		if m = strings.TrimSpace(m); m != "" {
			return m
		}
	}
	return DefaultModel
}

// validate rejects a value that would otherwise fall back to a default unnoticed.
func (c ChainsConfig) validate() error {
	models := map[string]string{"common": c.Common.Model, "docDefine": c.DocDefine.Model, "docEval": c.DocEval.Model, "docCreate": c.DocCreate.Model, "docUpdate": c.DocUpdate.Model}
	for name, m := range models {
		if m == "" {
			continue
		}
		if _, err := ValidateModel(m); err != nil {
			return fmt.Errorf("chains.%s.model: %w", name, err)
		}
	}
	strategies := map[string]string{"common": c.Common.Strategy, "docDefine": c.DocDefine.Strategy, "docEval": c.DocEval.Strategy}
	for name, s := range strategies {
		if s != "" && !ValidStrategy(s) {
			return fmt.Errorf("chains.%s.strategy must be %s, %s or %s, got %q", name, StrategyRandom, StrategyOldest, StrategyNewest, s)
		}
	}
	for name, i := range map[string]string{"docDefine": c.DocDefine.Interval, "docEval": c.DocEval.Interval} {
		if i == "" {
			continue
		}
		if _, _, err := ParseInterval(i); err != nil {
			return fmt.Errorf("chains.%s.interval: %w", name, err)
		}
	}
	return nil
}
