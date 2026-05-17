package models

type Binder struct {
	ID        int    `json:"id" gorm:"column:id"`
	Title     string `json:"title" gorm:"column:title"`
	Docs      int    `json:"docs" gorm:"column:docs"`
	Links     int    `json:"links" gorm:"column:links"`
	TitleDoc  string `json:"title_doc" gorm:"column:title_doc"`
	Enabled   bool   `json:"enabled" gorm:"column:enabled"`
	CreatedAt string `json:"created_at" gorm:"column:created_at"`
}
