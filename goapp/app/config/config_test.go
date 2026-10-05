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

func TestLoadRedisRoles(t *testing.T) {
	tests := []struct {
		name         string
		env          map[string]string
		wantPersist  RedisEndpoint
		wantVolatile RedisEndpoint
	}{
		{
			name:         "legacy variables are ignored",
			env:          map[string]string{"REDIS_HOST": "redis-cache", "REDIS_PORT": "6380"},
			wantPersist:  RedisEndpoint{Port: 6379},
			wantVolatile: RedisEndpoint{Port: 6379},
		},
		{
			name: "independent role endpoints",
			env: map[string]string{
				"REDIS_HOST":          "redis-cache",
				"REDIS_PERSIST_HOST":  "redis-persist",
				"REDIS_PERSIST_PORT":  "6382",
				"REDIS_VOLATILE_HOST": "redis-volatile",
				"REDIS_VOLATILE_PORT": "6381",
			},
			wantPersist:  RedisEndpoint{Host: "redis-persist", Port: 6382},
			wantVolatile: RedisEndpoint{Host: "redis-volatile", Port: 6381},
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
