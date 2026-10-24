package doceval

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
	docevalsvc "github.com/zetaoss/zengine/goapp/services/doceval"
)

// definitionItem is a stored definition: what the article is about, from its revision Revid.
type definitionItem struct {
	PageID       uint      `json:"page_id" gorm:"column:page_id"`
	PageTitle    string    `json:"page_title" gorm:"column:page_title"`
	Revid        uint64    `json:"revid" gorm:"column:revid"`
	Definition   string    `json:"definition" gorm:"column:definition"`
	Model        string    `json:"model" gorm:"column:model"`
	ChainVersion string    `json:"chain_version" gorm:"column:chain_version"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// Definitions lists stored definitions, latest first; filters q (title substring, spaces as underscores),
// limit (25, max 100), offset. Out-of-date ones are not marked: doc-define-loop redefines them on its own.
func Definitions(c *serverctx.Context) {
	query := c.R.URL.Query()
	q := c.DB.Model(&models.DocDefinition{})
	if title := strings.TrimSpace(query.Get("q")); title != "" {
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(title)
		q = q.Where("page_title LIKE ?", "%"+strings.ReplaceAll(escaped, " ", "\\_")+"%")
	}
	limit, offset := 25, 0
	if l, err := strconv.Atoi(query.Get("limit")); err == nil && l > 0 {
		limit = min(l, 100)
	}
	if o, err := strconv.Atoi(query.Get("offset")); err == nil && o >= 0 {
		offset = o
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		c.InternalError()
		return
	}
	items := []definitionItem{}
	if err := q.Order("updated_at DESC, page_id DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		c.InternalError()
		return
	}
	c.JSON(app.H{"total": total, "items": items, "limit": limit, "offset": offset})
}

// DefinitionCandidate returns the article doc-define-loop would define next by strategy (default: the setting).
func DefinitionCandidate(c *serverctx.Context) {
	strategy := strings.TrimSpace(c.R.URL.Query().Get("strategy"))
	if strategy == "" {
		if settings, err := docevalsvc.GetSettings(c.DB, c.Cfg); err == nil {
			strategy = settings.DefineStrategy
		}
	}
	candidate, err := docevalsvc.PickUndefined(c.R.Context(), c.DB, strategy)
	if err != nil {
		c.InternalError()
		return
	}
	if candidate == nil {
		c.JSONError(http.StatusNotFound, "every article is defined")
		return
	}
	c.JSON(app.H{
		"page_id":    candidate.PageID,
		"page_title": strings.ReplaceAll(candidate.PageTitle, "_", " "),
	})
}

// RunDefinition enqueues an immediate definition run even if the interval is off. A page given by page_id or
// page_title is redefined even when its definition is current; otherwise one is picked.
func RunDefinition(c *serverctx.Context) {
	runNow(c, docevalsvc.DefineLoopTaskType)
}
