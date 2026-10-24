package doceval

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
	docevalsvc "github.com/zetaoss/zengine/goapp/services/doceval"
)

// definitionItem is a stored definition with the article's state: Current is false when the article has a newer
// revision or title, or the definition is from another chain version, so the next doc-define-loop runs redefine it.
type definitionItem struct {
	PageID         uint      `json:"page_id" gorm:"column:page_id"`
	PageTitle      string    `json:"page_title" gorm:"column:page_title"`
	Revid          uint64    `json:"revid" gorm:"column:revid"`
	Definition     string    `json:"definition" gorm:"column:definition"`
	Model          string    `json:"model" gorm:"column:model"`
	ChainVersion   string    `json:"chain_version" gorm:"column:chain_version"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at"`
	LatestRevid    uint64    `json:"latest_revid" gorm:"column:latest_revid"`
	RevisionsAfter int       `json:"revisions_after" gorm:"column:revisions_after"`
	Current        bool      `json:"current" gorm:"column:current"`
}

// Definitions lists stored definitions, latest first; filters q (title substring, spaces as underscores),
// stale=1 (not current), limit (25, max 100), offset.
func Definitions(c *serverctx.Context) {
	query := c.R.URL.Query()
	where := []string{"1 = 1"}
	args := []any{docevalsvc.DefineVersionArg()}
	if q := strings.TrimSpace(query.Get("q")); q != "" {
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`).Replace(strings.ReplaceAll(q, " ", "_"))
		where = append(where, "d.page_title LIKE @title")
		args = append(args, sql.Named("title", "%"+escaped+"%"))
	}
	if query.Get("stale") == "1" {
		where = append(where, "NOT COALESCE("+docevalsvc.CurrentDefinitionSQL+", 0)")
	}
	limit, offset := 25, 0
	if l, err := strconv.Atoi(query.Get("limit")); err == nil && l > 0 {
		limit = min(l, 100)
	}
	if o, err := strconv.Atoi(query.Get("offset")); err == nil && o >= 0 {
		offset = o
	}
	from := `FROM doc_definitions d LEFT JOIN zetawiki.page p ON p.page_id = d.page_id WHERE ` + strings.Join(where, " AND ")

	var total int64
	if err := c.DB.Raw("SELECT COUNT(*) "+from, args...).Scan(&total).Error; err != nil {
		c.InternalError()
		return
	}
	items := []definitionItem{}
	err := c.DB.Raw(`
		SELECT d.page_id, d.page_title, d.revid, d.definition, d.model, d.chain_version, d.updated_at,
		  COALESCE(p.page_latest, 0) AS latest_revid,
		  (SELECT COUNT(*) FROM zetawiki.revision r WHERE r.rev_page = d.page_id AND r.rev_id > d.revid) AS revisions_after,
		  COALESCE(`+docevalsvc.CurrentDefinitionSQL+`, 0) AS current
		`+from+`
		ORDER BY d.updated_at DESC, d.page_id DESC
		LIMIT @limit OFFSET @offset
	`, append(args, sql.Named("limit", limit), sql.Named("offset", offset))...).Scan(&items).Error
	if err != nil {
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
