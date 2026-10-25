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

// definitionCurrent reports whether d was written for this revision and title by the current doc-define major.
func definitionCurrent(d *models.DocDefinition, page Page) bool {
	return d != nil && d.Revid == page.Revid && d.PageTitle == page.Title && prompts.SameMajor(d.ChainVersion, prompts.DocDefine.Metadata.Version)
}

// EnsureDefinition returns the page's definition, writing a new one with model (Define) when there is none or
// it is out of date (another revision, title, or major version).
func (e *Evaluator) EnsureDefinition(ctx context.Context, db *gorm.DB, page Page, wikitext, model string) (*models.DocDefinition, error) {
	d, err := GetDefinition(ctx, db, page.ID)
	if err != nil {
		return nil, fmt.Errorf("load definition: %w", err)
	}
	if definitionCurrent(d, page) {
		return d, nil
	}
	return e.Define(ctx, db, page, wikitext, model)
}

// Define writes the page's definition with model from its title and lead, and stores it in place of any
// earlier one. With EnsureDefinition it is the only writer of doc_definitions.
func (e *Evaluator) Define(ctx context.Context, db *gorm.DB, page Page, wikitext, model string) (*models.DocDefinition, error) {
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
	definition := truncateRunes(normalizeDefinition(page.Title, out.Definition), maxDefinitionLen)
	if definition == "" {
		return nil, fmt.Errorf("llm definition is empty")
	}

	d := &models.DocDefinition{
		PageID:       page.ID,
		PageTitle:    page.Title,
		Revid:        page.Revid,
		Definition:   definition,
		Model:        res.Model,
		ChainVersion: prompts.DocDefine.Metadata.Version,
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

// normalizeDefinition keeps a definition in the asked form, a noun phrase without the title, when the model
// answers with a sentence: it drops a leading title with its subject particle ("김치는 ") and a trailing
// copula ("음식이다.").
func normalizeDefinition(title, raw string) string {
	d := strings.Join(strings.Fields(raw), " ")
	display := strings.ReplaceAll(title, "_", " ")
	for _, particle := range []string{"은 ", "는 ", "이 ", "가 ", "란 ", "이란 "} {
		if rest, ok := strings.CutPrefix(d, display+particle); ok && rest != "" {
			d = rest
			break
		}
	}
	d = strings.TrimSuffix(d, ".")
	if rest, ok := strings.CutSuffix(d, "이다"); ok && rest != "" {
		d = rest
	}
	return strings.TrimSpace(d)
}

func modelOrDefault(model string) string {
	if m := strings.TrimSpace(model); m != "" {
		return m
	}
	return DefaultModel
}
