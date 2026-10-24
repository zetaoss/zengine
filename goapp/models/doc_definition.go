package models

import "time"

// DocDefinition is a one-line definition of what a wiki page is about, written by the doc-define chain
// from the page's title and lead at Revid. A page has one row, replaced when its revision or the chain
// version changes.
type DocDefinition struct {
	PageID       uint      `json:"page_id" gorm:"column:page_id;primaryKey;autoIncrement:false"`
	PageTitle    string    `json:"page_title" gorm:"column:page_title"`
	Revid        uint64    `json:"revid" gorm:"column:revid"`
	Definition   string    `json:"definition" gorm:"column:definition"`
	Model        string    `json:"model" gorm:"column:model"`
	ChainVersion string    `json:"chain_version" gorm:"column:chain_version"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (DocDefinition) TableName() string {
	return "doc_definitions"
}
