package doceval

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
	"github.com/zetaoss/zengine/goapp/prompts"
	"github.com/zetaoss/zengine/goapp/services/chainsvc"
	docevalsvc "github.com/zetaoss/zengine/goapp/services/doceval"
)

// DefineLoopTask writes the definition of one article per run: one without a current definition, picked by
// the define strategy, or one given in the payload. Evaluations pick only articles it has defined.
type DefineLoopTask struct{}

func NewDefineLoopTask() *DefineLoopTask {
	return &DefineLoopTask{}
}

// LoopInterval implements registry.Looper for self-perpetuating scheduling.
func (t *DefineLoopTask) LoopInterval(ctx context.Context, tc taskctx.Context) (time.Duration, bool, error) {
	s, err := loadSettings(tc)
	if err != nil {
		return 0, false, err
	}
	d, active := s.DefineDuration()
	return d, active, nil
}

// Execute defines a page given by p.PageID or p.PageTitle, always rewriting its definition, or else picks one
// without a current definition. p.Force runs it even when the definition interval is off.
func (t *DefineLoopTask) Execute(ctx context.Context, tc taskctx.Context, p LoopPayload) (app.H, error) {
	db, err := tc.GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get DB: %w", err)
	}
	settings, err := docevalsvc.GetSettings(db, tc.Config())
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}
	if _, active := settings.DefineDuration(); !active && !p.Force {
		return app.H{"skipped": true, "reason": "definition is off"}, nil
	}

	model := chainsvc.ModelFor(db, tc.Config(), prompts.DocDefine)
	run := docevalsvc.LastRun{At: time.Now(), Model: model}
	fail := func(err error) (app.H, error) {
		run.At, run.Error = time.Now(), err.Error()
		saveLastRun(db, docevalsvc.SettingKeyDefineLastRun, run)
		return nil, err
	}

	given := p.PageID > 0 || strings.TrimSpace(p.PageTitle) != ""
	var picked candidatePage
	if given {
		if picked, err = resolvePage(ctx, db, p.PageID, p.PageTitle); err != nil {
			return fail(err)
		}
	} else {
		c, err := docevalsvc.PickUndefined(ctx, db, settings.DefineStrategy)
		if err != nil {
			return fail(fmt.Errorf("failed to pick a page to define: %w", err))
		}
		if c == nil {
			return app.H{"skipped": true, "reason": "every article is defined"}, nil
		}
		picked = candidatePage(*c)
	}
	run.PageID, run.PageTitle = picked.PageID, picked.PageTitle

	wikitext, revid, err := fetchWikiPageContent(ctx, tc.Config(), picked.PageID)
	if err != nil {
		return fail(fmt.Errorf("failed to fetch wiki page: %w", err))
	}
	if revid == 0 {
		revid = picked.PageLatest
	}
	page := docevalsvc.Page{ID: uint(picked.PageID), Title: picked.PageTitle, Revid: uint64(revid)}
	evaluator := docevalsvc.New(tc.Config(), db)
	define := evaluator.EnsureDefinition
	if given {
		define = evaluator.Define
	}
	def, err := define(ctx, db, page, wikitext, model)
	if err != nil {
		return fail(fmt.Errorf("definition failed: %w", err))
	}

	run.At, run.OK, run.Model = time.Now(), true, def.Model
	saveLastRun(db, docevalsvc.SettingKeyDefineLastRun, run)
	slog.Info("[doc-define-loop] Defined page", "page_id", def.PageID, "title", def.PageTitle, "revid", def.Revid)
	return app.H{"ok": true, "page_id": def.PageID, "page_title": def.PageTitle, "definition": def.Definition}, nil
}
