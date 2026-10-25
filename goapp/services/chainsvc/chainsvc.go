// Package chainsvc keeps which model each built-in chain runs on. A sysop picks it in the chains tab; until then
// the GOAPP_CONFIG chains section decides, and then config.DefaultModel.
package chainsvc

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/prompts"
	"gorm.io/gorm"
)

// SettingKey is the app_settings row of the saved models: a JSON object of chain name to model.
const SettingKey = "chain_models"

// Sources of a chain's model.
const (
	SourceSaved   = "saved"
	SourceConfig  = "config"
	SourceDefault = "default"
)

// Saved returns the models a sysop saved, by chain name; none without a database.
func Saved(db *gorm.DB) (map[string]string, error) {
	if db == nil {
		return map[string]string{}, nil
	}
	var row models.AppSetting
	err := db.Table("app_settings").Where("setting_key = ?", SettingKey).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	saved := map[string]string{}
	if err := json.Unmarshal([]byte(row.ValueJSON), &saved); err != nil {
		return map[string]string{}, nil
	}
	return saved, nil
}

// Resolve returns the model of a chain given the saved models and where it came from: saved, the config file,
// or the default.
func Resolve(saved map[string]string, cfg *config.Config, chain string) (string, string) {
	if m := strings.TrimSpace(saved[chain]); m != "" {
		return m, SourceSaved
	}
	if cfg != nil {
		if m := cfg.Chains.ModelOf(chain); m != "" {
			return m, SourceConfig
		}
	}
	return config.DefaultModel, SourceDefault
}

// ModelFor returns the model a chain runs on. A failure to read the saved models falls back to the config file,
// so a run still has a model.
func ModelFor(db *gorm.DB, cfg *config.Config, chain prompts.Chain) string {
	saved, err := Saved(db)
	if err != nil {
		saved = nil
	}
	m, _ := Resolve(saved, cfg, chain.Name)
	return m
}

// Save sets the model of a built-in chain; an empty model removes the saved one, so the config file decides again.
// Each change is one statement on the chain's key of chain_models, so overlapping saves of different chains keep
// each other's models.
func Save(db *gorm.DB, chain, model string) error {
	if !builtin(chain) {
		return fmt.Errorf("unknown chain %q", chain)
	}
	model = strings.TrimSpace(model)
	if model != "" {
		var err error
		if model, err = config.ValidateModel(model); err != nil {
			return err
		}
	}
	// A built-in chain name has no quote, so it is safe inside the JSON path.
	path := `$."` + chain + `"`
	if model == "" {
		return db.Exec(`UPDATE app_settings SET value_json = JSON_REMOVE(value_json, ?) WHERE setting_key = ?`, path, SettingKey).Error
	}
	return db.Exec(`
		INSERT INTO app_settings (setting_key, value_json) VALUES (?, JSON_OBJECT(?, ?))
		ON DUPLICATE KEY UPDATE value_json = JSON_SET(value_json, ?, ?)
	`, SettingKey, chain, model, path, model).Error
}

func builtin(chain string) bool {
	for _, c := range prompts.BuiltinChains() {
		if c.Name == chain {
			return true
		}
	}
	return false
}
