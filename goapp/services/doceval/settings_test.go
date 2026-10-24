package doceval

import (
	"testing"
	"time"

	"github.com/zetaoss/zengine/goapp/app/config"
)

func TestParseInterval(t *testing.T) {
	tests := []struct {
		input       string
		expectedDur time.Duration
		expectedStr string
		expectErr   bool
	}{
		{"0s", 0, "0s", false},
		{"0", 0, "0s", false},
		{"Off", 0, "0s", false},
		{"off", 0, "0s", false},
		{"", 0, "0s", false},
		{"10m", 10 * time.Minute, "10m", false},
		{"20m", 20 * time.Minute, "20m", false},
		{"30m", 30 * time.Minute, "30m", false},
		{"1h", 1 * time.Hour, "1h", false},
		{"1h30m", 90 * time.Minute, "1h30m", false},
		{"6h", 6 * time.Hour, "6h", false},
		{"24h", 24 * time.Hour, "24h", false},
		{"99h", 99 * time.Hour, "99h", false},
		{"99h59m", 99*time.Hour + 59*time.Minute, "99h59m", false},
		{"100h", 0, "", true},    // must be < 100h
		{"150h", 0, "", true},    // must be < 100h
		{"500ms", 0, "", true},   // below MinActiveInterval (1s)
		{"invalid", 0, "", true}, // invalid format
	}

	for _, tc := range tests {
		dur, norm, err := ParseInterval(tc.input)
		if tc.expectErr {
			if err == nil {
				t.Errorf("input %q: expected error, got nil", tc.input)
			}
		} else {
			if err != nil {
				t.Errorf("input %q: unexpected error: %v", tc.input, err)
			}
			if dur != tc.expectedDur || norm != tc.expectedStr {
				t.Errorf("input %q: got (%v, %q), want (%v, %q)", tc.input, dur, norm, tc.expectedDur, tc.expectedStr)
			}
		}
	}
}

func TestDefaultSettingsBuiltin(t *testing.T) {
	s := DefaultSettings(nil)
	if d, active := s.Duration(); !active || d != 1*time.Hour {
		t.Fatalf("default evaluation interval must be 1h active, got %v active=%v", d, active)
	}
	if d, active := s.DefineDuration(); !active || d != 30*time.Minute {
		t.Fatalf("default definition interval must be 30m active, got %v active=%v", d, active)
	}
	if s.Strategy != StrategyRandom || s.DefineStrategy != StrategyRandom {
		t.Fatalf("unexpected defaults %+v", s)
	}
}

func TestDefaultSettingsFromConfig(t *testing.T) {
	cfg := &config.Config{Chains: config.ChainsConfig{
		Common:    config.CommonChainConfig{Model: "gemini/gemini-3.8-flash", Strategy: "Newest"},
		DocDefine: config.ChainConfig{Interval: "10m"},
		DocEval:   config.ChainConfig{Interval: "0s"},
	}}
	s := DefaultSettings(cfg)
	if s.Interval != "0s" || s.Strategy != StrategyNewest {
		t.Fatalf("expected docEval over common, got %+v", s)
	}
	if s.DefineInterval != "10m" || s.DefineStrategy != StrategyNewest {
		t.Fatalf("expected docDefine over common, got %+v", s)
	}
	bad := &config.Config{Chains: config.ChainsConfig{DocEval: config.ChainConfig{Interval: "500ms"}}}
	if got := DefaultSettings(bad).Interval; got != "1h" {
		t.Fatalf("expected an invalid configured interval to fall back, got %q", got)
	}
}

func TestSaveSettingsRejectsInvalidInterval(t *testing.T) {
	// Validation runs before the database is touched, so a nil DB is enough.
	for _, s := range []Settings{{Interval: "500ms"}, {Interval: "0s", DefineInterval: "200h"}} {
		if err := SaveSettings(nil, s); err == nil {
			t.Errorf("expected an error for %+v", s)
		}
	}
}
