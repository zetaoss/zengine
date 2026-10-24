package doceval

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
	"github.com/zetaoss/zengine/goapp/prompts"
	docevalsvc "github.com/zetaoss/zengine/goapp/services/doceval"
	"gorm.io/gorm"
)

type candidatePage struct {
	PageID     int    `gorm:"column:page_id"`
	PageTitle  string `gorm:"column:page_title"`
	PageLatest int    `gorm:"column:page_latest"`
}

type LoopPayload struct {
	Force     bool   `json:"force,omitempty"`
	PageID    int    `json:"page_id,omitempty"`
	PageTitle string `json:"page_title,omitempty"`
}

var wikiSpaceRegex = regexp.MustCompile(`[\s_]+`)

// normalizeWikiTitle normalizes a user-supplied title into MediaWiki canonical form:
// trims surrounding whitespace and underscores, collapses consecutive whitespace and
// underscores into a single underscore, and capitalizes the first unicode character.
func normalizeWikiTitle(title string) string {
	title = strings.Trim(title, " \t\r\n_")
	if title == "" {
		return ""
	}
	title = wikiSpaceRegex.ReplaceAllString(title, "_")
	runes := []rune(title)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// LoopTask is the worker task that evaluates ZetaWiki articles.
type LoopTask struct{}

func NewLoopTask() *LoopTask {
	return &LoopTask{}
}

// LoopInterval implements registry.Looper for self-perpetuating scheduling.
func (t *LoopTask) LoopInterval(ctx context.Context, tc taskctx.Context) (time.Duration, bool, error) {
	s, err := loadSettings(tc)
	if err != nil {
		return 0, false, err
	}
	d, active := s.Duration()
	return d, active, nil
}

func loadSettings(tc taskctx.Context) (docevalsvc.Settings, error) {
	db, err := tc.GetDB()
	if err != nil {
		return docevalsvc.Settings{}, err
	}
	return docevalsvc.GetSettings(db, tc.Config())
}

func (t *LoopTask) Execute(ctx context.Context, taskCtx taskctx.Context, p LoopPayload) (app.H, error) {
	db, err := taskCtx.GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get DB: %w", err)
	}

	settings, err := docevalsvc.GetSettings(db, taskCtx.Config())
	if err != nil {
		slog.Warn("[doc-eval-loop] Failed to load settings, falling back to defaults", "err", err)
		settings = docevalsvc.DefaultSettings(taskCtx.Config())
	}

	if _, active := settings.Duration(); !active && !p.Force {
		slog.Info("[doc-eval-loop] Evaluation is currently Off (0s). Skipping.")
		return app.H{"skipped": true, "reason": "evaluation is off"}, nil
	}

	modelToRun := strings.TrimSpace(settings.Model)
	if modelToRun == "" {
		modelToRun = docevalsvc.DefaultModel
	}

	var candidate candidatePage

	// fail records the failed run for the settings panel. The registry framework handles
	// loop perpetuation automatically.
	fail := func(err error) (app.H, error) {
		saveLastRun(db, docevalsvc.SettingKeyDocEvalLastRun, docevalsvc.LastRun{At: time.Now(), PageID: candidate.PageID, PageTitle: candidate.PageTitle, Model: modelToRun, Error: err.Error()})
		return nil, err
	}
	if p.PageID > 0 || strings.TrimSpace(p.PageTitle) != "" {
		if candidate, err = resolvePage(ctx, db, p.PageID, p.PageTitle); err != nil {
			return fail(err)
		}
	} else {
		picked, err := docevalsvc.PickCandidate(ctx, db, settings.Strategy)
		if err != nil {
			return fail(fmt.Errorf("failed to pick candidate page: %w", err))
		}
		if picked != nil {
			candidate = candidatePage(*picked)
		}
	}

	if candidate.PageID == 0 {
		slog.Warn("[doc-eval-loop] No candidate page found for evaluation")
		return app.H{"skipped": true, "reason": "no candidate found"}, nil
	}

	slog.Info("[doc-eval-loop] Picked candidate for evaluation", "page_id", candidate.PageID, "title", candidate.PageTitle)

	cfg := taskCtx.Config()
	wikitext, revid, err := fetchWikiPageContent(ctx, cfg, candidate.PageID)
	if err != nil {
		return fail(fmt.Errorf("failed to fetch wiki page: %w", err))
	}
	if revid == 0 && candidate.PageLatest > 0 {
		revid = candidate.PageLatest
	}

	// An auto-picked page already has a current definition; a page given by title or ID, or one edited since
	// it was picked, gets one written here.
	evaluator := docevalsvc.New(cfg, db)
	page := docevalsvc.Page{ID: uint(candidate.PageID), Title: candidate.PageTitle, Revid: uint64(revid)}
	def, err := evaluator.EnsureDefinition(ctx, db, page, wikitext, settings.DefineModel)
	if err != nil {
		return fail(fmt.Errorf("definition failed: %w", err))
	}

	var cachedRubrics string
	_ = db.WithContext(ctx).Table("doc_evaluations").
		Where("page_id = ? AND page_title = ? AND chain_version = ? AND definition = ?", candidate.PageID, candidate.PageTitle, prompts.DocEval.Metadata["version"], def.Definition).
		Order("id DESC").Limit(1).Pluck("rubrics", &cachedRubrics).Error

	evalRecord, err := evaluator.Evaluate(ctx, page, def.Definition, wikitext, settings.Model, cachedRubrics)
	if err != nil {
		return fail(fmt.Errorf("evaluation failed: %w", err))
	}

	if err := db.WithContext(ctx).Create(evalRecord).Error; err != nil {
		return fail(fmt.Errorf("failed to save doc evaluation: %w", err))
	}
	saveLastRun(db, docevalsvc.SettingKeyDocEvalLastRun, docevalsvc.LastRun{At: time.Now(), OK: true, PageID: candidate.PageID, PageTitle: candidate.PageTitle, Model: evalRecord.Model, EvalID: evalRecord.ID})

	slog.Info("[doc-eval-loop] Evaluation completed successfully",
		"id", evalRecord.ID,
		"page_id", evalRecord.PageID,
		"title", evalRecord.PageTitle,
		"grade", evalRecord.Grade,
		"general_quality", evalRecord.GeneralQuality,
		"elapsed_ms", evalRecord.ElapsedMs,
	)

	return app.H{
		"ok":              true,
		"eval_id":         evalRecord.ID,
		"page_id":         evalRecord.PageID,
		"page_title":      evalRecord.PageTitle,
		"grade":           evalRecord.Grade,
		"general_quality": evalRecord.GeneralQuality,
	}, nil
}

// resolvePage finds the article for a page ID or a title, following a redirect; a title is normalized and
// matched case-insensitively as a last resort.
func resolvePage(ctx context.Context, db *gorm.DB, pageID int, pageTitle string) (candidatePage, error) {
	var candidate candidatePage
	if pageID > 0 {
		query := `
			SELECT p.page_id, p.page_title, p.page_latest
			FROM zetawiki.page p
			WHERE p.page_id = ? AND p.page_namespace = 0 AND p.page_is_redirect = 0
			LIMIT 1
		`
		if err := db.WithContext(ctx).Raw(query, pageID).Scan(&candidate).Error; err != nil {
			return candidate, fmt.Errorf("failed to find page ID %d: %w", pageID, err)
		}
		if candidate.PageID == 0 {
			// Resolve redirect if page_id points to a redirect page
			redirectQuery := `
				SELECT target.page_id, target.page_title, target.page_latest
				FROM zetawiki.page src
				JOIN zetawiki.redirect r ON r.rd_from = src.page_id
				JOIN zetawiki.page target ON target.page_namespace = r.rd_namespace AND target.page_title = r.rd_title
				WHERE src.page_id = ? AND target.page_namespace = 0 AND target.page_is_redirect = 0
				LIMIT 1
			`
			if err := db.WithContext(ctx).Raw(redirectQuery, pageID).Scan(&candidate).Error; err != nil {
				return candidate, fmt.Errorf("failed to resolve redirect for page ID %d: %w", pageID, err)
			}
		}
		if candidate.PageID == 0 {
			return candidate, fmt.Errorf("page ID %d not found or not an article", pageID)
		}
	} else if title := strings.TrimSpace(pageTitle); title != "" {
		normalized := normalizeWikiTitle(title)
		query := `
			SELECT p.page_id, p.page_title, p.page_latest
			FROM zetawiki.page p
			WHERE p.page_namespace = 0 AND p.page_is_redirect = 0 AND p.page_title = ?
			LIMIT 1
		`
		if err := db.WithContext(ctx).Raw(query, normalized).Scan(&candidate).Error; err != nil {
			return candidate, fmt.Errorf("failed to find page %q: %w", title, err)
		}
		if candidate.PageID == 0 {
			// Resolve redirect if the title points to a redirect page
			redirectQuery := `
				SELECT target.page_id, target.page_title, target.page_latest
				FROM zetawiki.page src
				JOIN zetawiki.redirect r ON r.rd_from = src.page_id
				JOIN zetawiki.page target ON target.page_namespace = r.rd_namespace AND target.page_title = r.rd_title
				WHERE src.page_namespace = 0 AND src.page_title = ? AND target.page_namespace = 0 AND target.page_is_redirect = 0
				LIMIT 1
			`
			if err := db.WithContext(ctx).Raw(redirectQuery, normalized).Scan(&candidate).Error; err != nil {
				return candidate, fmt.Errorf("failed to resolve redirect for page %q: %w", title, err)
			}
		}
		if candidate.PageID == 0 {
			// Case-insensitive fallback for direct match in namespace 0
			ciQuery := `
				SELECT p.page_id, p.page_title, p.page_latest
				FROM zetawiki.page p
				WHERE p.page_namespace = 0 AND p.page_is_redirect = 0 AND LOWER(CONVERT(p.page_title USING utf8mb4)) = LOWER(?)
				LIMIT 1
			`
			_ = db.WithContext(ctx).Raw(ciQuery, normalized).Scan(&candidate).Error
		}
		if candidate.PageID == 0 {
			return candidate, fmt.Errorf("page not found or not an article: %q", title)
		}
	}
	return candidate, nil
}

func saveLastRun(db *gorm.DB, key string, r docevalsvc.LastRun) {
	if err := docevalsvc.SaveLastRun(db, key, r); err != nil {
		slog.Warn("[doc-eval] Failed to save last run", "key", key, "err", err)
	}
}

func fetchWikiPageContent(ctx context.Context, cfg *config.Config, pageID int) (string, int, error) {
	apiServer := "http://127.0.0.1"
	if cfg != nil && cfg.App.APIServer != "" {
		apiServer = cfg.App.APIServer
	}
	url := fmt.Sprintf("%s/w/api.php?action=query&prop=revisions&pageids=%d&rvprop=ids|content&rvslots=main&format=json", apiServer, pageID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", 0, fmt.Errorf("create mw request: %w", err)
	}
	req.Header.Set("User-Agent", "ZetaEngine-DocEval/1.0")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("fetch mw content: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("mw api status: %d", resp.StatusCode)
	}

	var raw struct {
		Query struct {
			Pages json.RawMessage `json:"pages"`
		} `json:"query"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return "", 0, fmt.Errorf("decode mw api response: %w", err)
	}

	type mwPage struct {
		PageID    int  `json:"pageid"`
		Missing   bool `json:"missing"`
		Redirect  bool `json:"redirect"`
		Revisions []struct {
			RevID int `json:"revid"`
			Slots struct {
				Main struct {
					Content string `json:"content"`
					Star    string `json:"*"`
				} `json:"main"`
			} `json:"slots"`
		} `json:"revisions"`
	}

	var pagesList []mwPage
	var pagesMap map[string]mwPage
	if err := json.Unmarshal(raw.Query.Pages, &pagesMap); err == nil && len(pagesMap) > 0 {
		for _, p := range pagesMap {
			pagesList = append(pagesList, p)
		}
	} else {
		if err := json.Unmarshal(raw.Query.Pages, &pagesList); err != nil {
			return "", 0, fmt.Errorf("decode mw api pages: %w", err)
		}
	}

	if len(pagesList) == 0 || pagesList[0].Missing {
		return "", 0, fmt.Errorf("page %d not found or missing", pageID)
	}

	page := pagesList[0]
	if page.Redirect {
		return "", 0, fmt.Errorf("page %d is a redirect", pageID)
	}

	if len(page.Revisions) == 0 {
		return "", 0, fmt.Errorf("page %d has no revisions", pageID)
	}

	content := page.Revisions[0].Slots.Main.Content
	if content == "" {
		content = page.Revisions[0].Slots.Main.Star
	}

	return content, page.Revisions[0].RevID, nil
}
