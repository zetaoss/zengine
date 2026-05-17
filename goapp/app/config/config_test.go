package config

import (
	"testing"
)

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("DB_HOST", "db.local")
	t.Setenv("DB_PORT", "3307")
	t.Setenv("DEV_MODE", "true")
	t.Setenv("AD_SLOTS", "a,b")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.DB.Host != "db.local" {
		t.Fatalf("Common.DBHost = %q, want db.local", cfg.DB.Host)
	}
	if cfg.DB.Port != 3307 {
		t.Fatalf("Common.DBPort = %d, want 3307", cfg.DB.Port)
	}
	if !cfg.App.DevMode {
		t.Fatalf("Server.DevMode = %t, want true", cfg.App.DevMode)
	}
	if len(cfg.Ads.Slots) != 2 || cfg.Ads.Slots[0] != "a" || cfg.Ads.Slots[1] != "b" {
		t.Fatalf("Server.AdSlots = %#v, want [a b]", cfg.Ads.Slots)
	}
}
