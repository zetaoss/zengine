package aimodels

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/prompts"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
	"github.com/zetaoss/zengine/goapp/services/llmsvc/client"
	"gorm.io/gorm"
)

// ModelHistory is zengine's record of one model over all its chain calls (llm_calls).
type ModelHistory struct {
	Model         string     `json:"model" gorm:"column:model"`
	Calls         int64      `json:"calls" gorm:"column:calls"`
	OKs           int64      `json:"oks" gorm:"column:oks"`
	AvgElapsedMs  float64    `json:"avg_elapsed_ms" gorm:"column:avg_elapsed_ms"`
	FirstCallAt   *time.Time `json:"first_call_at" gorm:"column:first_call_at"`
	LastOKAt      *time.Time `json:"last_ok_at" gorm:"column:last_ok_at"`
	LastFailureAt *time.Time `json:"last_failure_at" gorm:"column:last_failure_at"`
	LastError     string     `json:"last_error" gorm:"column:last_error"`
	Chains        string     `json:"chains" gorm:"column:chains"`
}

// Index returns bob aigate's models per provider (all available, with enabled flags), today's usage of the
// configured models (null when bob has no /v1/usage), and zengine's own call history per model.
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
	history, err := modelHistory(c.DB)
	if err != nil {
		slog.Warn("[aimodels] Failed to load model history", "err", err)
	}
	c.JSON(app.H{"models": models, "usage": usage, "history": history})
}

// modelHistory aggregates llm_calls per requested model, most calls first, with the error of its latest failure.
func modelHistory(db *gorm.DB) ([]ModelHistory, error) {
	history := []ModelHistory{}
	err := db.Raw(`
		SELECT c.model,
		  COUNT(*) AS calls,
		  SUM(c.ok) AS oks,
		  AVG(c.elapsed_ms) AS avg_elapsed_ms,
		  MIN(c.created_at) AS first_call_at,
		  MAX(CASE WHEN c.ok = 1 THEN c.created_at END) AS last_ok_at,
		  MAX(CASE WHEN c.ok = 0 THEN c.created_at END) AS last_failure_at,
		  COALESCE((SELECT f.error FROM llm_calls f WHERE f.model = c.model AND f.ok = 0 ORDER BY f.id DESC LIMIT 1), '') AS last_error,
		  GROUP_CONCAT(DISTINCT c.chain ORDER BY c.chain SEPARATOR ',') AS chains
		FROM llm_calls c
		GROUP BY c.model
		ORDER BY calls DESC
	`).Scan(&history).Error
	return history, err
}

// BuiltinChains lists the built-in prompt chains kept in code (goapp/prompts), read-only.
func BuiltinChains(c *serverctx.Context) {
	c.JSON(prompts.BuiltinChains())
}
