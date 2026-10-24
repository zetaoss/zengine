package doceval

import (
	"testing"
	"time"
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

func TestDefaultSettingsInterval(t *testing.T) {
	if d, active := DefaultSettings().Duration(); !active || d != 1*time.Hour {
		t.Fatalf("default settings must be 1h active, got %v active=%v", d, active)
	}
}

func TestDefaultSettingsModel(t *testing.T) {
	if got := DefaultSettings().Model; got != "groq/openai/gpt-oss-120b" {
		t.Fatalf("unexpected default model %q", got)
	}
}

func TestSaveSettingsRejectsInvalidModel(t *testing.T) {
	// Validation runs before the database is touched, so a nil DB is enough.
	for _, model := range []string{"", "gemini", "/x", "gemini/"} {
		if err := SaveSettings(nil, Settings{Interval: "0s", Model: model}); err == nil {
			t.Errorf("expected an error for model %q", model)
		}
	}
}
