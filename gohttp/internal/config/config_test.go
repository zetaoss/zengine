package config

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseMode(t *testing.T) {
	t.Parallel()

	t.Run("default prod", func(t *testing.T) {
		t.Parallel()
		got, err := parseMode(nil)
		if err != nil {
			t.Fatalf("parseMode error: %v", err)
		}
		if got != "prod" {
			t.Fatalf("mode = %q, want %q", got, "prod")
		}
	})

	t.Run("dev true", func(t *testing.T) {
		t.Parallel()
		got, err := parseMode([]string{"--dev=true"})
		if err != nil {
			t.Fatalf("parseMode error: %v", err)
		}
		if got != "dev" {
			t.Fatalf("mode = %q, want %q", got, "dev")
		}
	})

	t.Run("reject positional args", func(t *testing.T) {
		t.Parallel()
		_, err := parseMode([]string{"extra"})
		if err == nil {
			t.Fatal("expected error for positional args, got nil")
		}
		if !strings.Contains(err.Error(), "unexpected positional args") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestParseCommaSeparated(t *testing.T) {
	t.Parallel()

	got := parseCommaSeparated("slotA, slotB ,slotC")
	want := []string{"slotA", "slotB", "slotC"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("parts[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	empty := parseCommaSeparated("")
	if len(empty) != 0 {
		t.Fatalf("parseCommaSeparated empty = %v, want []", empty)
	}

	filtered := parseCommaSeparated("slotA, , ,slotB,,   ")
	filteredWant := []string{"slotA", "slotB"}
	if len(filtered) != len(filteredWant) {
		t.Fatalf("filtered len = %d, want %d (%v)", len(filtered), len(filteredWant), filtered)
	}
	for i := range filteredWant {
		if filtered[i] != filteredWant[i] {
			t.Fatalf("filtered[%d] = %q, want %q", i, filtered[i], filteredWant[i])
		}
	}
}

func TestZConfStaticFromEnv(t *testing.T) {
	t.Setenv("AVATAR_BASE_URL", "https://avatar.example")
	t.Setenv("GA_MEASUREMENT_ID", "G-TEST123")
	t.Setenv("AD_CLIENT", "ca-pub-123")
	t.Setenv("AD_SLOTS", "foo,bar")

	got, err := zConfStaticFromEnv()
	if err != nil {
		t.Fatalf("zConfStaticFromEnv error: %v", err)
	}

	finalJSON := `{` + got + `,"policy":"strict"}`
	if !json.Valid([]byte(finalJSON)) {
		t.Fatalf("injected zconf is not valid json: %q", finalJSON)
	}

	checks := []string{
		`"avatarBaseUrl":"https://avatar.example"`,
		`"gaMeasurementId":"G-TEST123"`,
		`"adClient":"ca-pub-123"`,
		`"adSlots":["foo","bar"]`,
	}
	for _, s := range checks {
		if !strings.Contains(got, s) {
			t.Fatalf("zconf static missing %q in %q", s, got)
		}
	}
}
