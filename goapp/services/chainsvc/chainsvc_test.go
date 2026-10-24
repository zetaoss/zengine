package chainsvc

import (
	"testing"

	"github.com/zetaoss/zengine/goapp/app/config"
)

func TestResolve(t *testing.T) {
	cfg := &config.Config{Chains: config.ChainsConfig{Common: config.CommonChainConfig{Model: "gemini/a"}, DocEval: config.ChainConfig{Model: "groq/b"}}}
	saved := map[string]string{"doc-define": "cloudflare/c"}
	for chain, want := range map[string][2]string{
		"doc-define": {"cloudflare/c", SourceSaved},
		"doc-eval":   {"groq/b", SourceConfig},
		"doc-create": {"gemini/a", SourceConfig},
	} {
		if m, src := Resolve(saved, cfg, chain); m != want[0] || src != want[1] {
			t.Errorf("%s: got %s %s, want %v", chain, m, src, want)
		}
	}
	if m, src := Resolve(nil, nil, "doc-update"); m != config.DefaultModel || src != SourceDefault {
		t.Errorf("default: got %s %s", m, src)
	}
}

func TestSaveRejects(t *testing.T) {
	// Validation runs before the database is touched, so a nil DB is enough.
	if err := Save(nil, "no-such-chain", "groq/a"); err == nil {
		t.Error("expected an unknown chain to be rejected")
	}
	if err := Save(nil, "doc-eval", "gpt"); err == nil {
		t.Error("expected a model without provider to be rejected")
	}
}
