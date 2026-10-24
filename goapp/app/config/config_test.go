package config

import (
	"os"
	"testing"
)

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("DB_HOST", "db.local")
	t.Setenv("DEV_MODE", "true")
	t.Setenv("AD_SLOTS", "a,b")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.DB.Host != "db.local" {
		t.Fatalf("Common.DBHost = %q, want db.local", cfg.DB.Host)
	}
	if !cfg.App.DevMode {
		t.Fatalf("Server.DevMode = %t, want true", cfg.App.DevMode)
	}
	if len(cfg.Ads.Slots) != 2 || cfg.Ads.Slots[0] != "a" || cfg.Ads.Slots[1] != "b" {
		t.Fatalf("Server.AdSlots = %#v, want [a b]", cfg.Ads.Slots)
	}
}

// window.ZCONF.adSlots must be an array ([]), not null, when AD_SLOTS is empty.
func TestLoadEmptyAdSlots(t *testing.T) {
	t.Setenv("AD_SLOTS", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.Ads.Slots == nil || len(cfg.Ads.Slots) != 0 {
		t.Fatalf("Ads.Slots = %#v, want empty non-nil slice", cfg.Ads.Slots)
	}
}

func TestLoadRedisRoles(t *testing.T) {
	tests := []struct {
		name         string
		env          map[string]string
		wantPersist  RedisEndpoint
		wantVolatile RedisEndpoint
	}{
		{
			name: "legacy variables are ignored",
			env: map[string]string{
				"REDIS_HOST":          "redis-cache",
				"REDIS_PORT":          "6380",
				"REDIS_PERSIST_PORT":  "6381",
				"REDIS_VOLATILE_PORT": "6382",
			},
		},
		{
			name: "independent role hosts; ports are fixed",
			env: map[string]string{
				"REDIS_HOST":          "redis-cache",
				"REDIS_PERSIST_HOST":  "redis-persist",
				"REDIS_PERSIST_PORT":  "6382",
				"REDIS_VOLATILE_HOST": "redis-volatile",
				"REDIS_VOLATILE_PORT": "6381",
			},
			wantPersist:  RedisEndpoint{Host: "redis-persist"},
			wantVolatile: RedisEndpoint{Host: "redis-volatile"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range []string{"REDIS_HOST", "REDIS_PORT", "REDIS_PERSIST_HOST", "REDIS_PERSIST_PORT", "REDIS_VOLATILE_HOST", "REDIS_VOLATILE_PORT"} {
				t.Setenv(k, "")
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			cfg, err := Load()
			if err != nil {
				t.Fatalf("Load error: %v", err)
			}
			if cfg.Redis.Persist != tt.wantPersist {
				t.Fatalf("Redis.Persist = %+v, want %+v", cfg.Redis.Persist, tt.wantPersist)
			}
			if cfg.Redis.Volatile != tt.wantVolatile {
				t.Fatalf("Redis.Volatile = %+v, want %+v", cfg.Redis.Volatile, tt.wantVolatile)
			}
		})
	}
}

func TestLoadChainsFromFile(t *testing.T) {
	path := t.TempDir() + "/goapp.yaml"
	if err := os.WriteFile(path, []byte(`chains:
  common:
    interval: 1h
    model: groq/openai/gpt-oss-120b
    strategy: random
  docDefine:
    interval: 30m
  docEval:
    # interval: 1h
`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOAPP_CONFIG", path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	def := cfg.Chains.DocDefine.Resolve(cfg.Chains.Common)
	eval := cfg.Chains.DocEval.Resolve(cfg.Chains.Common)
	if def != (ChainConfig{Interval: "30m", Model: "groq/openai/gpt-oss-120b", Strategy: "random"}) {
		t.Fatalf("docDefine = %+v", def)
	}
	if eval != (ChainConfig{Interval: "1h", Model: "groq/openai/gpt-oss-120b", Strategy: "random"}) {
		t.Fatalf("docEval = %+v", eval)
	}
}

func TestLoadFileErrors(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/typo.yaml", []byte("chains:\n  docEvel:\n    interval: 1h\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for name, path := range map[string]string{"missing file": dir + "/none.yaml", "unknown key": dir + "/typo.yaml"} {
		t.Setenv("GOAPP_CONFIG", path)
		if _, err := Load(); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}

	if err := os.WriteFile(dir+"/empty.yaml", nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOAPP_CONFIG", dir+"/empty.yaml")
	if _, err := Load(); err != nil {
		t.Errorf("an empty file must load: %v", err)
	}
}
