package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	App        AppConfig
	DB         DBConfig
	Redis      RedisConfig
	Ads        AdsConfig
	Analytics  AnalyticsConfig
	API        APIConfig
	Cloudflare CloudflareConfig
	OAuth      OAuthConfig
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
	Port     int
	Database string
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
	GAPropertyID    string
	GAReaderFile    string
	GATimezone      string
	GSCSiteURL      string
}

type APIConfig struct {
	MonitoringEndpoint  string
	MonitoringNamespace string
	MonitoringNodepool  string
	MonitoringPVC       string
	LLMEndpoint         string
	RunboxEndpoint      string
	SearchEndpoint      string
}

type CloudflareConfig struct {
	APIToken string
	ZoneID   string
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
		App:        AppConfig{},
		DB:         DBConfig{},
		Redis:      RedisConfig{},
		Ads:        AdsConfig{},
		Analytics:  AnalyticsConfig{},
		API:        APIConfig{},
		Cloudflare: CloudflareConfig{},
		OAuth:      OAuthConfig{},
	}

	cfg.App.APIServer = lookup("API_SERVER")
	cfg.App.AppURL = lookup("APP_URL")
	cfg.App.AvatarBaseURL = lookup("AVATAR_BASE_URL")
	cfg.App.DevMode = lookupBool("DEV_MODE", false)
	cfg.App.InternalSecretKey = lookup("INTERNAL_SECRET_KEY")
	cfg.App.LogLevel = lookupString("LOG_LEVEL", "info")

	cfg.DB.Host = lookup("DB_HOST")
	cfg.DB.Port = lookupInt("DB_PORT", 3306)
	cfg.DB.Database = lookup("DB_DATABASE")
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
	cfg.Analytics.GAPropertyID = lookup("GA_PROPERTY_ID")
	cfg.Analytics.GAReaderFile = lookup("GA_READER_FILE")
	cfg.Analytics.GATimezone = lookup("GA_TIMEZONE")
	cfg.Analytics.GSCSiteURL = lookup("GSC_SITE_URL")

	cfg.API.MonitoringEndpoint = lookup("MONITORING_ENDPOINT")
	cfg.API.MonitoringNamespace = lookup("MONITORING_NAMESPACE")
	cfg.API.MonitoringNodepool = lookup("MONITORING_NODEPOOL")
	cfg.API.MonitoringPVC = lookup("MONITORING_PVC")
	cfg.API.LLMEndpoint = lookup("LLM_ENDPOINT")
	cfg.API.RunboxEndpoint = lookup("RUNBOX_ENDPOINT")
	cfg.API.SearchEndpoint = lookup("SEARCH_ENDPOINT")

	cfg.Cloudflare.APIToken = lookup("CLOUDFLARE_API_TOKEN")
	cfg.Cloudflare.ZoneID = lookup("CLOUDFLARE_ZONE_ID")

	cfg.OAuth.FacebookClientID = lookup("FACEBOOK_CLIENT_ID")
	cfg.OAuth.FacebookClientSecret = lookup("FACEBOOK_CLIENT_SECRET")
	cfg.OAuth.GithubClientID = lookup("GITHUB_CLIENT_ID")
	cfg.OAuth.GithubClientSecret = lookup("GITHUB_CLIENT_SECRET")
	cfg.OAuth.GoogleClientID = lookup("GOOGLE_CLIENT_ID")
	cfg.OAuth.GoogleClientSecret = lookup("GOOGLE_CLIENT_SECRET")

	return cfg, nil
}

func lookup(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func lookupInt(key string, def int) int {
	raw := lookup(key)
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return n
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
		return nil
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
