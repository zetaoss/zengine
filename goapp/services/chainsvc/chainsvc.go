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
	"gorm.io/gorm/clause"
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
	saved, err := Saved(db)
	if err != nil {
		return err
	}
	if model == "" {
		delete(saved, chain)
	} else {
		saved[chain] = model
	}
	raw, err := json.Marshal(saved)
	if err != nil {
		return err
	}
	return db.Table("app_settings").Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "setting_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value_json"}),
	}).Create(&models.AppSetting{SettingKey: SettingKey, ValueJSON: string(raw)}).Error
}

func builtin(chain string) bool {
	for _, c := range prompts.BuiltinChains() {
		if c.Name == chain {
			return true
		}
	}
	return false
}
