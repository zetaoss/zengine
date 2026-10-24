package models

import (
	"time"

	"gorm.io/gorm"
)

// DocEvalGradeFail is the grade of a document that violates a safety policy; other grades are S, A, B, C, D
// by composite score (80% general quality + 20% text quality). Contradicted sentences (AccuracyContradicted)
// are shown beside the grade, not in it.
const DocEvalGradeFail = "F"

// CompositeScore calculates the weighted overall score: 80% general quality + 20% text quality.
func CompositeScore(generalQuality, textQuality uint8) uint8 {
	return uint8((uint(generalQuality)*8 + uint(textQuality)*2 + 5) / 10)
}

type DocEvaluation struct {
	ID                   uint64    `json:"id" gorm:"column:id;primaryKey"`
	PageID               uint      `json:"page_id" gorm:"column:page_id"`
	PageTitle            string    `json:"page_title" gorm:"column:page_title"`
	Revid                uint64    `json:"revid" gorm:"column:revid"`
	Definition           string    `json:"definition" gorm:"column:definition"`
	Grade                string    `json:"grade" gorm:"column:grade"`
	GeneralQuality       uint8     `json:"general_quality" gorm:"column:general_quality"`
	TextQuality          uint8     `json:"text_quality" gorm:"column:text_quality"`
	Score                uint8     `json:"score" gorm:"-"`
	SafetyViolated       bool      `json:"safety_violated" gorm:"column:safety_violated"`
	SafetyPolicies       string    `json:"safety_policies" gorm:"column:safety_policies"`
	AccuracyContradicted uint      `json:"accuracy_contradicted" gorm:"column:accuracy_contradicted"`
	Summary              string    `json:"summary" gorm:"column:summary"`
	Rubrics              string    `json:"rubrics" gorm:"column:rubrics"`
	ChainVersion         string    `json:"chain_version" gorm:"column:chain_version"`
	RawJSON              string    `json:"raw_json" gorm:"column:raw_json"`
	Model                string    `json:"model" gorm:"column:model"`
	PromptTokens         uint      `json:"prompt_tokens" gorm:"column:prompt_tokens"`
	CompletionTokens     uint      `json:"completion_tokens" gorm:"column:completion_tokens"`
	ElapsedMs            uint      `json:"elapsed_ms" gorm:"column:elapsed_ms"`
	CreatedAt            time.Time `json:"created_at" gorm:"column:created_at"`

	// Enriched fields for display (not stored in doc_evaluations table)
	LatestRevid    uint64 `json:"latest_revid" gorm:"-"`
	IsLatest       bool   `json:"is_latest" gorm:"-"`
	RevisionsAfter int    `json:"revisions_after" gorm:"-"`
}

func (d *DocEvaluation) AfterFind(tx *gorm.DB) error {
	d.Score = CompositeScore(d.GeneralQuality, d.TextQuality)
	return nil
}

func (DocEvaluation) TableName() string {
	return "doc_evaluations"
}
