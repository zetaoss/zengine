package config

import (
	"bufio"
	"fmt"
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
	EditBot    EditBotConfig
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

type RedisConfig struct {
	Host string
	Port int
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
	LLMEndpoint    string
	RunboxEndpoint string
	SearchEndpoint string
}

type CloudflareConfig struct {
	APIToken string
	ZoneID   string
}

type EditBotConfig struct {
	Username string
	Password string
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
	overrides := map[string]string{}
	envFilePath := "/app/.env"
	if _, err := os.Stat(envFilePath); err == nil {
		parsed, err := parseEnvFile(envFilePath)
		if err != nil {
			return nil, err
		}
		overrides = parsed
	}

	cfg := &Config{
		App:        AppConfig{},
		DB:         DBConfig{},
		Redis:      RedisConfig{},
		Ads:        AdsConfig{},
		Analytics:  AnalyticsConfig{},
		API:        APIConfig{},
		Cloudflare: CloudflareConfig{},
		EditBot:    EditBotConfig{},
		OAuth:      OAuthConfig{},
	}

	cfg.App.APIServer = lookup(overrides, "API_SERVER")
	cfg.App.AppURL = lookup(overrides, "APP_URL")
	cfg.App.AvatarBaseURL = lookup(overrides, "AVATAR_BASE_URL")
	cfg.App.DevMode = lookupBool(overrides, "DEV_MODE", false)
	cfg.App.InternalSecretKey = lookup(overrides, "INTERNAL_SECRET_KEY")
	cfg.App.LogLevel = lookupString(overrides, "LOG_LEVEL", "info")

	cfg.DB.Host = lookup(overrides, "DB_HOST")
	cfg.DB.Port = lookupInt(overrides, "DB_PORT", 3306)
	cfg.DB.Database = lookup(overrides, "DB_DATABASE")
	cfg.DB.Username = lookup(overrides, "DB_USERNAME")
	cfg.DB.Password = lookup(overrides, "DB_PASSWORD")

	cfg.Redis.Host = lookup(overrides, "REDIS_HOST")
	cfg.Redis.Port = lookupInt(overrides, "REDIS_PORT", 6379)

	cfg.Ads.Client = lookup(overrides, "AD_CLIENT")
	cfg.Ads.Slots = lookupList(overrides, "AD_SLOTS")

	cfg.Analytics.GAMeasurementID = lookup(overrides, "GA_MEASUREMENT_ID")
	cfg.Analytics.GAPropertyID = lookup(overrides, "GA_PROPERTY_ID")
	cfg.Analytics.GAReaderFile = lookup(overrides, "GA_READER_FILE")
	cfg.Analytics.GATimezone = lookup(overrides, "GA_TIMEZONE")
	cfg.Analytics.GSCSiteURL = lookup(overrides, "GSC_SITE_URL")

	cfg.API.LLMEndpoint = lookup(overrides, "LLM_ENDPOINT")
	cfg.API.RunboxEndpoint = lookup(overrides, "RUNBOX_ENDPOINT")
	cfg.API.SearchEndpoint = lookup(overrides, "SEARCH_ENDPOINT")

	cfg.Cloudflare.APIToken = lookup(overrides, "CLOUDFLARE_API_TOKEN")
	cfg.Cloudflare.ZoneID = lookup(overrides, "CLOUDFLARE_ZONE_ID")

	cfg.EditBot.Username = lookup(overrides, "EDITBOT_USERNAME")
	cfg.EditBot.Password = lookup(overrides, "EDITBOT_PASSWORD")

	cfg.OAuth.FacebookClientID = lookup(overrides, "FACEBOOK_CLIENT_ID")
	cfg.OAuth.FacebookClientSecret = lookup(overrides, "FACEBOOK_CLIENT_SECRET")
	cfg.OAuth.GithubClientID = lookup(overrides, "GITHUB_CLIENT_ID")
	cfg.OAuth.GithubClientSecret = lookup(overrides, "GITHUB_CLIENT_SECRET")
	cfg.OAuth.GoogleClientID = lookup(overrides, "GOOGLE_CLIENT_ID")
	cfg.OAuth.GoogleClientSecret = lookup(overrides, "GOOGLE_CLIENT_SECRET")

	return cfg, nil
}

func lookup(overrides map[string]string, key string) string {
	if v, ok := overrides[key]; ok {
		return strings.TrimSpace(v)
	}
	return strings.TrimSpace(os.Getenv(key))
}

func lookupInt(overrides map[string]string, key string, def int) int {
	raw := lookup(overrides, key)
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return n
}

func lookupString(overrides map[string]string, key string, def string) string {
	raw := lookup(overrides, key)
	if raw == "" {
		return def
	}
	return raw
}

func lookupBool(overrides map[string]string, key string, def bool) bool {
	raw := strings.ToLower(lookup(overrides, key))
	switch raw {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}

func lookupList(overrides map[string]string, key string) []string {
	raw := lookup(overrides, key)
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

func parseEnvFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open env file %s: %w", path, err)
	}
	defer func() {
		_ = f.Close()
	}()

	out := map[string]string{}
	sc := bufio.NewScanner(f)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		i := strings.IndexRune(line, '=')
		if i < 1 {
			return nil, fmt.Errorf("invalid env line %d in %s", lineNo, path)
		}
		key := strings.TrimSpace(line[:i])
		val := strings.TrimSpace(line[i+1:])
		val = trimEnvQuotes(val)
		out[key] = val
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read env file %s: %w", path, err)
	}
	return out, nil
}

func trimEnvQuotes(v string) string {
	if len(v) >= 2 {
		if (v[0] == '\'' && v[len(v)-1] == '\'') || (v[0] == '"' && v[len(v)-1] == '"') {
			return v[1 : len(v)-1]
		}
	}
	return v
}
