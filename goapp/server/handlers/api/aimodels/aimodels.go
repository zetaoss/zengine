package aimodels

import (
	"encoding/json"
	"net/http"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/prompts"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
	"github.com/zetaoss/zengine/goapp/services/llmsvc/client"
)

// Index returns bob aigate's models per provider (all available, with enabled flags) and today's usage
// of the configured models. usage is null when bob has no /v1/usage.
func Index(c *serverctx.Context) {
	llm := client.New(c.Cfg)

	models, status, err := llm.GetJSON(c.R.Context(), "/v1/models?all")
	if err != nil || status != http.StatusOK {
		c.JSONError(http.StatusBadGateway, "failed to load models from bob")
		return
	}
	var usage json.RawMessage
	if body, status, err := llm.GetJSON(c.R.Context(), "/v1/usage"); err == nil && status == http.StatusOK {
		usage = body
	}
	c.JSON(app.H{"models": models, "usage": usage})
}

// BuiltinChains lists the built-in prompt chains kept in code (goapp/prompts), read-only.
func BuiltinChains(c *serverctx.Context) {
	c.JSON(prompts.BuiltinChains())
}
