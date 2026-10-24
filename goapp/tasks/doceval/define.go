package doceval

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
	docevalsvc "github.com/zetaoss/zengine/goapp/services/doceval"
)

// DefineLoopTask writes the definition of one article per run: one without a current definition, picked by
// the define strategy. Evaluations pick only articles it has defined.
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

func (t *DefineLoopTask) Execute(ctx context.Context, tc taskctx.Context, _ struct{}) (app.H, error) {
	db, err := tc.GetDB()
	if err != nil {
		return nil, fmt.Errorf("failed to get DB: %w", err)
	}
	settings, err := docevalsvc.GetSettings(db, tc.Config())
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}
	if _, active := settings.DefineDuration(); !active {
		return app.H{"skipped": true, "reason": "definition is off"}, nil
	}

	picked, err := docevalsvc.PickUndefined(ctx, db, settings.DefineStrategy)
	if err != nil {
		return nil, fmt.Errorf("failed to pick a page to define: %w", err)
	}
	if picked == nil {
		return app.H{"skipped": true, "reason": "every article is defined"}, nil
	}

	run := docevalsvc.LastRun{At: time.Now(), PageID: picked.PageID, PageTitle: picked.PageTitle, Model: settings.DefineModel}
	fail := func(err error) (app.H, error) {
		run.At, run.Error = time.Now(), err.Error()
		saveLastRun(db, docevalsvc.SettingKeyDefineLastRun, run)
		return nil, err
	}
	wikitext, revid, err := fetchWikiPageContent(ctx, tc.Config(), picked.PageID)
	if err != nil {
		return fail(fmt.Errorf("failed to fetch wiki page: %w", err))
	}
	if revid == 0 {
		revid = picked.PageLatest
	}
	page := docevalsvc.Page{ID: uint(picked.PageID), Title: picked.PageTitle, Revid: uint64(revid)}
	def, err := docevalsvc.New(tc.Config(), db).EnsureDefinition(ctx, db, page, wikitext, settings.DefineModel)
	if err != nil {
		return fail(fmt.Errorf("definition failed: %w", err))
	}

	run.At, run.OK, run.Model = time.Now(), true, def.Model
	saveLastRun(db, docevalsvc.SettingKeyDefineLastRun, run)
	slog.Info("[doc-define-loop] Defined page", "page_id", def.PageID, "title", def.PageTitle, "revid", def.Revid)
	return app.H{"ok": true, "page_id": def.PageID, "page_title": def.PageTitle, "definition": def.Definition}, nil
}
