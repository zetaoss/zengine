package doceval

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/models"
	"github.com/zetaoss/zengine/goapp/prompts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// maxLeadLen is how much of the wikitext the define step reads: the lead says what the page is about.
const maxLeadLen = 1500

// maxDefinitionLen matches doc_definitions.definition and doc_evaluations.definition.
const maxDefinitionLen = 500

// Page is a revision of a wiki article to define or evaluate. Title is the stored DB key (Linux_명령어).
type Page struct {
	ID    uint
	Title string
	Revid uint64
}

// GetDefinition returns the stored definition of a page, or nil when there is none.
func GetDefinition(ctx context.Context, db *gorm.DB, pageID uint) (*models.DocDefinition, error) {
	var d models.DocDefinition
	err := db.WithContext(ctx).Where("page_id = ?", pageID).Take(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// definitionCurrent reports whether d was written for this revision and title by the current doc-define chain.
func definitionCurrent(d *models.DocDefinition, page Page) bool {
	return d != nil && d.Revid == page.Revid && d.PageTitle == page.Title && d.ChainVersion == prompts.DocDefine.Metadata["version"]
}

// EnsureDefinition returns the page's definition, writing and storing a new one with model when there is none
// or it is out of date (another revision, title, or chain version). It is the only writer of doc_definitions.
func (e *Evaluator) EnsureDefinition(ctx context.Context, db *gorm.DB, page Page, wikitext, model string) (*models.DocDefinition, error) {
	d, err := GetDefinition(ctx, db, page.ID)
	if err != nil {
		return nil, fmt.Errorf("load definition: %w", err)
	}
	if definitionCurrent(d, page) {
		return d, nil
	}

	lead := wikitext
	if len(lead) > maxLeadLen {
		lead = strings.ToValidUTF8(lead[:maxLeadLen], "")
	}
	var out struct {
		Definition string `json:"definition"`
	}
	res, err := e.llm.RunStep(ctx, modelOrDefault(model), prompts.DocDefine, prompts.DocDefineStep, map[string]string{
		"title": strings.ReplaceAll(page.Title, "_", " "),
		"lead":  lead,
	}, &out)
	if err != nil {
		return nil, fmt.Errorf("llm definition failed: %w", err)
	}
	definition := truncateRunes(strings.Join(strings.Fields(out.Definition), " "), maxDefinitionLen)
	if definition == "" {
		return nil, fmt.Errorf("llm definition is empty")
	}

	d = &models.DocDefinition{
		PageID:       page.ID,
		PageTitle:    page.Title,
		Revid:        page.Revid,
		Definition:   definition,
		Model:        res.Model,
		ChainVersion: prompts.DocDefine.Metadata["version"],
		UpdatedAt:    time.Now(),
	}
	err = db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "page_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"page_title", "revid", "definition", "model", "chain_version", "updated_at"}),
	}).Create(d).Error
	if err != nil {
		return nil, fmt.Errorf("save definition: %w", err)
	}
	return d, nil
}

func modelOrDefault(model string) string {
	if m := strings.TrimSpace(model); m != "" {
		return m
	}
	return DefaultModel
}
