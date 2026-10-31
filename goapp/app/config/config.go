package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App       AppConfig
	DB        DBConfig
	Redis     RedisConfig
	Ads       AdsConfig
	Analytics AnalyticsConfig
	API       APIConfig
	OAuth     OAuthConfig
	Chains    ChainsConfig
}

type AppConfig struct {
	APIServer         string
	AppURL            string
	AvatarBaseURL     string
	DevMode           bool
	InternalSecretKey string
	LogLevel          string
}

type DBConfig struct {
	Host     string
	Username string
	Password string
}

// RedisConfig holds the two Redis roles.
// Persist must not evict or lose data: task queues, auth tokens, rate-limit counters.
// Volatile may evict anything: caches.
type RedisConfig struct {
	Persist  RedisEndpoint
	Volatile RedisEndpoint
}

type RedisEndpoint struct {
	Host string
}

type AdsConfig struct {
	Client string
	Slots  []string
}

type AnalyticsConfig struct {
	GAMeasurementID string
}

type APIConfig struct {
	// BobEndpoint is the base URL of bob, the in-cluster app server. Callers append the
	// feature path: /aigate (LLM), /cloudflare, /ga, /gsc (analytics), /metrics (k8s stats), /runbox, /search.
	BobEndpoint string
}

// fileConfig is the GOAPP_CONFIG file: settings that are not secrets or per-environment endpoints.
type fileConfig struct {
	Chains ChainsConfig `yaml:"chains"`
}

type OAuthConfig struct {
	FacebookClientID     string
	FacebookClientSecret string
	GithubClientID       string
	GithubClientSecret   string
	GoogleClientID       string
	GoogleClientSecret   string
}

func Load() (*Config, error) {
	cfg := &Config{
		App:       AppConfig{},
		DB:        DBConfig{},
		Redis:     RedisConfig{},
		Ads:       AdsConfig{},
		Analytics: AnalyticsConfig{},
		API:       APIConfig{},
		OAuth:     OAuthConfig{},
	}

	cfg.App.APIServer = lookup("API_SERVER")
	cfg.App.AppURL = lookup("APP_URL")
	cfg.App.AvatarBaseURL = lookup("AVATAR_BASE_URL")
	cfg.App.DevMode = lookupBool("DEV_MODE", false)
	cfg.App.InternalSecretKey = lookup("INTERNAL_SECRET_KEY")
	cfg.App.LogLevel = lookupString("LOG_LEVEL", "info")

	cfg.DB.Host = lookup("DB_HOST")
	cfg.DB.Username = lookup("DB_USERNAME")
	cfg.DB.Password = lookup("DB_PASSWORD")

	// Each role has its own host; the Redis port is fixed at 6379.
	cfg.Redis.Persist = RedisEndpoint{
		Host: lookup("REDIS_PERSIST_HOST"),
	}
	cfg.Redis.Volatile = RedisEndpoint{
		Host: lookup("REDIS_VOLATILE_HOST"),
	}

	cfg.Ads.Client = lookup("AD_CLIENT")
	cfg.Ads.Slots = lookupList("AD_SLOTS")

	cfg.Analytics.GAMeasurementID = lookup("GA_MEASUREMENT_ID")

	cfg.API.BobEndpoint = strings.TrimRight(lookup("BOB_ENDPOINT"), "/")

	cfg.OAuth.FacebookClientID = lookup("FACEBOOK_CLIENT_ID")
	cfg.OAuth.FacebookClientSecret = lookup("FACEBOOK_CLIENT_SECRET")
	cfg.OAuth.GithubClientID = lookup("GITHUB_CLIENT_ID")
	cfg.OAuth.GithubClientSecret = lookup("GITHUB_CLIENT_SECRET")
	cfg.OAuth.GoogleClientID = lookup("GOOGLE_CLIENT_ID")
	cfg.OAuth.GoogleClientSecret = lookup("GOOGLE_CLIENT_SECRET")

	if path := lookup("GOAPP_CONFIG"); path != "" {
		file, err := loadFile(path)
		if err != nil {
			return nil, err
		}
		if err := file.Chains.validate(); err != nil {
			return nil, fmt.Errorf("GOAPP_CONFIG %s: %w", path, err)
		}
		cfg.Chains = file.Chains
	}

	return cfg, nil
}

// loadFile reads the GOAPP_CONFIG file. Unknown keys are errors, so a typo does not silently keep a default.
func loadFile(path string) (fileConfig, error) {
	var file fileConfig
	raw, err := os.ReadFile(path)
	if err != nil {
		return file, fmt.Errorf("read GOAPP_CONFIG %s: %w", path, err)
	}
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(&file); err != nil && !errors.Is(err, io.EOF) {
		return file, fmt.Errorf("parse GOAPP_CONFIG %s: %w", path, err)
	}
	return file, nil
}

func lookup(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func lookupString(key string, def string) string {
	raw := lookup(key)
	if raw == "" {
		return def
	}
	return raw
}

func lookupBool(key string, def bool) bool {
	raw := strings.ToLower(lookup(key))
	switch raw {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}

func lookupList(key string) []string {
	raw := lookup(key)
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}
