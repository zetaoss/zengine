package aimodels

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/prompts"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
	"github.com/zetaoss/zengine/goapp/services/chainsvc"
	"github.com/zetaoss/zengine/goapp/services/llmsvc/client"
	"gorm.io/gorm"
)

// ModelHistory is zengine's record of one model over all its chain calls (llm_calls).
type ModelHistory struct {
	Model         string         `json:"model" gorm:"column:model"`
	Calls         int64          `json:"calls" gorm:"column:calls"`
	OKs           int64          `json:"oks" gorm:"column:oks"`
	AvgElapsedMs  float64        `json:"avg_elapsed_ms" gorm:"column:avg_elapsed_ms"`
	FirstOKAt     *time.Time     `json:"first_ok_at" gorm:"column:first_ok_at"`
	LastOKAt      *time.Time     `json:"last_ok_at" gorm:"column:last_ok_at"`
	LastFailureAt *time.Time     `json:"last_failure_at" gorm:"column:last_failure_at"`
	LastError     string         `json:"last_error" gorm:"column:last_error"`
	Chains        []ChainHistory `json:"chains" gorm:"-"`
}

// ChainHistory is a model's calls from one chain.
type ChainHistory struct {
	Chain string `json:"chain" gorm:"column:chain"`
	Calls int64  `json:"calls" gorm:"column:calls"`
	OKs   int64  `json:"oks" gorm:"column:oks"`
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

// modelHistory aggregates llm_calls per requested model, most calls first, with the error of its latest failure
// and the calls per chain.
func modelHistory(db *gorm.DB) ([]ModelHistory, error) {
	history := []ModelHistory{}
	err := db.Raw(`
		SELECT c.model,
		  COUNT(*) AS calls,
		  SUM(c.ok) AS oks,
		  AVG(c.elapsed_ms) AS avg_elapsed_ms,
		  MIN(CASE WHEN c.ok = 1 THEN c.created_at END) AS first_ok_at,
		  MAX(CASE WHEN c.ok = 1 THEN c.created_at END) AS last_ok_at,
		  MAX(CASE WHEN c.ok = 0 THEN c.created_at END) AS last_failure_at,
		  COALESCE((SELECT f.error FROM llm_calls f WHERE f.model = c.model AND f.ok = 0 ORDER BY f.id DESC LIMIT 1), '') AS last_error
		FROM llm_calls c
		GROUP BY c.model
		ORDER BY calls DESC
	`).Scan(&history).Error
	if err != nil || len(history) == 0 {
		return history, err
	}
	var rows []struct {
		Model string `gorm:"column:model"`
		ChainHistory
	}
	if err := db.Raw(`
		SELECT model, chain, COUNT(*) AS calls, SUM(ok) AS oks
		FROM llm_calls
		GROUP BY model, chain
		ORDER BY calls DESC
	`).Scan(&rows).Error; err != nil {
		return history, err
	}
	byModel := make(map[string]int, len(history))
	for i := range history {
		byModel[history[i].Model] = i
		history[i].Chains = []ChainHistory{}
	}
	for _, r := range rows {
		if i, ok := byModel[r.Model]; ok {
			history[i].Chains = append(history[i].Chains, r.ChainHistory)
		}
	}
	return history, nil
}

// chainView is a built-in chain with the model it runs on.
type chainView struct {
	prompts.Chain
	Model string `json:"model"`
}

// Chains lists the built-in chains (goapp/prompts), each with its model.
func Chains(c *serverctx.Context) {
	saved, err := chainsvc.Saved(c.DB)
	if err != nil {
		c.InternalError()
		return
	}
	views := []chainView{}
	for _, chain := range prompts.BuiltinChains() {
		model, _ := chainsvc.Resolve(saved, c.Cfg, chain.Name)
		views = append(views, chainView{Chain: chain, Model: model})
	}
	c.JSON(views)
}

// SetChainModel sets the model of a built-in chain from {"model": "provider/model"}.
func SetChainModel(c *serverctx.Context) {
	name := strings.TrimSpace(c.R.PathValue("name"))
	var body struct {
		Model string `json:"model"`
	}
	if !c.Decode(&body) {
		return
	}
	if err := chainsvc.Save(c.DB, name, body.Model); err != nil {
		c.JSONError(http.StatusUnprocessableEntity, err.Error())
		return
	}
	saved, err := chainsvc.Saved(c.DB)
	if err != nil {
		c.InternalError()
		return
	}
	model, _ := chainsvc.Resolve(saved, c.Cfg, name)
	c.JSON(app.H{"ok": true, "chain": name, "model": model})
}
