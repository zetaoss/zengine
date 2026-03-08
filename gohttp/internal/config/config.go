package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type Config struct {
	Mode        string
	ZConfStatic string
}

func Load(args []string) (Config, error) {
	mode, err := parseMode(args)
	if err != nil {
		return Config{}, err
	}

	zConfStatic, err := zConfStaticFromEnv()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Mode:        mode,
		ZConfStatic: zConfStatic,
	}, nil
}

func parseMode(args []string) (string, error) {
	fs := flag.NewFlagSet("gohttp", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	dev := fs.Bool("dev", false, "run in dev mode")
	if err := fs.Parse(args); err != nil {
		return "", fmt.Errorf("usage: gohttp [--dev=true|false]")
	}
	if fs.NArg() > 0 {
		return "", fmt.Errorf("unexpected positional args: %v (usage: gohttp [--dev=true|false])", fs.Args())
	}
	if *dev {
		return "dev", nil
	}
	return "prod", nil
}

func zConfStaticFromEnv() (string, error) {
	payload, err := json.Marshal(zConf{
		AvatarBaseURL:   os.Getenv("AVATAR_BASE_URL"),
		GAMeasurementID: os.Getenv("GA_MEASUREMENT_ID"),
		AdClient:        os.Getenv("AD_CLIENT"),
		AdSlots:         parseCommaSeparated(os.Getenv("AD_SLOTS")),
	})
	if err != nil {
		return "", fmt.Errorf("build zconf: %w", err)
	}
	if len(payload) < 2 || payload[0] != '{' || payload[len(payload)-1] != '}' {
		return "", fmt.Errorf("invalid zconf payload")
	}
	// Return only inner fields; injector wraps them in window.ZCONF={...}.
	return string(payload[1 : len(payload)-1]), nil
}

type zConf struct {
	AvatarBaseURL   string   `json:"avatarBaseUrl"`
	GAMeasurementID string   `json:"gaMeasurementId"`
	AdClient        string   `json:"adClient"`
	AdSlots         []string `json:"adSlots"`
}

func parseCommaSeparated(raw string) []string {
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
