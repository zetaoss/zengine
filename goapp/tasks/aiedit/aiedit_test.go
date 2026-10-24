package aiedit

import (
	"testing"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/models"
)

func TestChainFor(t *testing.T) {
	cfg := &config.Config{Chains: config.ChainsConfig{
		Common:    config.CommonChainConfig{Model: "gemini/a"},
		DocUpdate: config.ModelChainConfig{Model: "groq/b"},
	}}
	if chain, model := chainFor(cfg, models.AIEditRequestUpdate); chain.Name != "doc-update" || model != "groq/b" {
		t.Errorf("update: got %s %s", chain.Name, model)
	}
	if chain, model := chainFor(cfg, models.AIEditRequestCreate); chain.Name != "doc-create" || model != "gemini/a" {
		t.Errorf("create: got %s %s", chain.Name, model)
	}
	if _, model := chainFor(nil, models.AIEditRequestCreate); model != config.DefaultModel {
		t.Errorf("no config: got %s", model)
	}
}
