package doceval

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/server/serverctx"
	docevalsvc "github.com/zetaoss/zengine/goapp/services/doceval"
	"gorm.io/gorm"
)

// All evaluation routes are sysop-only (r.Sysop() in routes.go).
func Index(c *serverctx.Context) {
	q := c.DB.Table("doc_evaluations")

	if pageIDStr := c.R.URL.Query().Get("page_id"); pageIDStr != "" {
		if pid, err := strconv.Atoi(pageIDStr); err == nil && pid > 0 {
			q = q.Where("page_id = ?", pid)
		}
	}

	if grade := strings.TrimSpace(c.R.URL.Query().Get("grade")); grade != "" {
		q = q.Where("grade = ?", grade)
	}
	if c.R.URL.Query().Get("contradicted") == "1" {
		q = q.Where("accuracy_contradicted > 0")
	}

	if maxStr := c.R.URL.Query().Get("max_quality"); maxStr != "" {
		if v, err := strconv.Atoi(maxStr); err == nil {
			q = q.Where("general_quality <= ?", v)
		}
	}
	if minStr := c.R.URL.Query().Get("min_quality"); minStr != "" {
		if v, err := strconv.Atoi(minStr); err == nil {
			q = q.Where("general_quality >= ?", v)
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		c.InternalError()
		return
	}

	limit := 25
	if limitStr := c.R.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > 100 {
				l = 100
			}
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.R.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var items []models.DocEvaluation
	if err := q.Order("id DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		c.InternalError()
		return
	}

	populateRevisionsInfo(c.DB, items)

	c.JSON(app.H{
		"total":  total,
		"items":  items,
		"limit":  limit,
		"offset": offset,
	})
}

func Show(c *serverctx.Context) {
	id, ok := c.PathInt("id")
	if !ok || id <= 0 {
		c.JSONError(http.StatusNotFound, "not found")
		return
	}

	var item models.DocEvaluation
	if err := c.DB.Table("doc_evaluations").Where("id = ?", id).Take(&item).Error; err != nil {
		c.JSONError(http.StatusNotFound, "not found")
		return
	}

	items := []models.DocEvaluation{item}
	populateRevisionsInfo(c.DB, items)
	item = items[0]

	c.JSON(item)
}

// populateRevisionsInfo adds each evaluation's page_latest and the number of edits made after the
// evaluated revision, in one query.
func populateRevisionsInfo(db *gorm.DB, items []models.DocEvaluation) {
	if len(items) == 0 || db == nil {
		return
	}
	ids := make([]uint64, len(items))
	for i, it := range items {
		ids[i] = it.ID
	}

	var rows []struct {
		ID             uint64 `gorm:"column:id"`
		PageLatest     uint64 `gorm:"column:page_latest"`
		RevisionsAfter int    `gorm:"column:revisions_after"`
	}
	err := db.Raw(`
		SELECT e.id, p.page_latest,
		  (SELECT COUNT(*) FROM zetawiki.revision r WHERE r.rev_page = e.page_id AND r.rev_id > e.revid) AS revisions_after
		FROM doc_evaluations e
		JOIN zetawiki.page p ON p.page_id = e.page_id
		WHERE e.id IN ? AND e.revid > 0
	`, ids).Scan(&rows).Error
	if err != nil {
		slog.Warn("[doc-eval] Failed to query revision status", "err", err)
		return
	}
	byID := make(map[uint64]int, len(rows))
	for i, r := range rows {
		byID[r.ID] = i
	}
	for i := range items {
		if j, ok := byID[items[i].ID]; ok {
			items[i].LatestRevid = rows[j].PageLatest
			items[i].RevisionsAfter = rows[j].RevisionsAfter
			items[i].IsLatest = items[i].Revid == rows[j].PageLatest
		}
	}
}

func Summary(c *serverctx.Context) {
	type gradeStat struct {
		Grade string `json:"grade"`
		Count int64  `json:"count"`
	}
	var gradeStats []gradeStat
	var stats struct {
		Total             int64   `gorm:"column:total"`
		Contradicted      int64   `gorm:"column:contradicted"`
		AvgGeneralQuality float64 `gorm:"column:avg_general_quality"`
		AvgTextQuality    float64 `gorm:"column:avg_text_quality"`
	}
	err := c.DB.Table("doc_evaluations").Select("grade, count(*) as count").Group("grade").Scan(&gradeStats).Error
	if err == nil {
		err = c.DB.Table("doc_evaluations").Select(`COUNT(*) AS total,
			COALESCE(SUM(accuracy_contradicted > 0), 0) AS contradicted,
			COALESCE(AVG(general_quality), 0) AS avg_general_quality,
			COALESCE(AVG(text_quality), 0) AS avg_text_quality`).Scan(&stats).Error
	}
	if err != nil {
		slog.Error("[doc-eval] Failed to summarize evaluations", "err", err)
		c.InternalError()
		return
	}

	c.JSON(app.H{
		"total_evaluations":   stats.Total,
		"avg_general_quality": stats.AvgGeneralQuality,
		"avg_text_quality":    stats.AvgTextQuality,
		"grades":              gradeStats,
		"contradicted":        stats.Contradicted,
	})
}

func GetSettings(c *serverctx.Context) {
	settings, err := docevalsvc.GetSettings(c.DB)
	if err != nil {
		c.InternalError()
		return
	}
	lastRun, _ := docevalsvc.GetLastRun(c.DB)
	queued, err := docevalsvc.QueuedLoopTasks(c.Cfg)
	if err != nil {
		slog.Warn("[doc-eval] Failed to list loop tasks", "err", err)
	}
	c.JSON(app.H{"interval": settings.Interval, "model": settings.Model, "strategy": settings.Strategy, "last_run": lastRun, "tasks": queued})
}

// updateSettingsReq updates the interval, the model, the strategy, or any combination; an omitted field keeps its value.
type updateSettingsReq struct {
	Interval *string `json:"interval"`
	Model    *string `json:"model"`
	Strategy *string `json:"strategy"`
}

func UpdateSettings(c *serverctx.Context) {
	var req updateSettingsReq
	if err := json.NewDecoder(c.R.Body).Decode(&req); err != nil {
		c.JSONError(http.StatusBadRequest, "invalid request body")
		return
	}

	settings, err := docevalsvc.GetSettings(c.DB)
	if err != nil {
		c.InternalError()
		return
	}
	var dur time.Duration
	if req.Interval != nil {
		d, norm, err := docevalsvc.ParseInterval(*req.Interval)
		if err != nil {
			c.JSONError(http.StatusBadRequest, err.Error())
			return
		}
		dur, settings.Interval = d, norm
	}
	if req.Model != nil {
		settings.Model = *req.Model
	}
	if req.Strategy != nil {
		settings.Strategy = docevalsvc.NormalizeStrategy(*req.Strategy)
	}
	if err := docevalsvc.SaveSettings(c.DB, settings); err != nil {
		c.JSONError(http.StatusBadRequest, err.Error())
		return
	}

	// A new interval restarts the loop: its waiting run moves to one interval from now, or is removed when off.
	if req.Interval != nil {
		var err error
		if dur > 0 {
			err = docevalsvc.ScheduleLoop(c.R.Context(), c.Cfg, c, dur)
		} else {
			err = docevalsvc.StopLoop(c.R.Context(), c.Cfg)
		}
		if err != nil {
			slog.Error("[doc-eval] Failed to reschedule loop after settings update", "err", err)
			c.JSONError(http.StatusInternalServerError, "settings saved, but failed to reschedule the loop: "+err.Error())
			return
		}
	}

	c.JSON(app.H{
		"ok":       true,
		"interval": settings.Interval,
		"model":    settings.Model,
		"strategy": settings.Strategy,
	})
}

type runReq struct {
	PageID    int    `json:"page_id"`
	PageTitle string `json:"page_title"`
}

// Candidate returns a candidate article title according to strategy (defaults to current setting).
func Candidate(c *serverctx.Context) {
	strategy := strings.TrimSpace(c.R.URL.Query().Get("strategy"))
	if strategy == "" {
		settings, err := docevalsvc.GetSettings(c.DB)
		if err == nil {
			strategy = settings.Strategy
		}
	}
	candidate, err := docevalsvc.PickCandidate(c.R.Context(), c.DB, strategy)
	if err != nil {
		c.InternalError()
		return
	}
	if candidate == nil {
		c.JSONError(http.StatusNotFound, "no article found")
		return
	}
	c.JSON(app.H{
		"page_id":    candidate.PageID,
		"page_title": strings.ReplaceAll(candidate.PageTitle, "_", " "),
	})
}

// Run enqueues a one-off evaluation run immediately, even if the interval is off.
// If page_title is supplied, that page is evaluated; otherwise a candidate is picked.
// The run replaces the waiting scheduled run, so the next one is scheduled from this execution
// (like Grafana auto-refresh reset).
func Run(c *serverctx.Context) {
	var req runReq
	if c.R.Body != nil {
		_ = json.NewDecoder(c.R.Body).Decode(&req)
	}

	info, err := docevalsvc.RunNow(c.R.Context(), c.Cfg, c, req.PageID, req.PageTitle)
	if info == nil {
		c.JSONError(http.StatusInternalServerError, "failed to enqueue task: "+err.Error())
		return
	}
	if err != nil {
		// The run is enqueued; the leftover waiting run is dropped when the loop reschedules.
		slog.Warn("[doc-eval] Failed to replace waiting loop task before manual run", "err", err)
	}
	c.JSON(app.H{
		"ok":      true,
		"task_id": info.ID,
	})
}
