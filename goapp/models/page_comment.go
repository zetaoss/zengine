package models

import "time"

type PageComment struct {
	ID       int       `json:"id" gorm:"column:id"`
	PageID   int       `json:"page_id" gorm:"column:page_id"`
	UserID   int       `json:"user_id" gorm:"column:user_id"`
	UserName string    `json:"user_name" gorm:"column:user_name"`
	Created  time.Time `json:"created" gorm:"column:created"`
	Message  string    `json:"message" gorm:"column:message"`
}

func (PageComment) TableName() string {
	return "zetawiki.page_comments"
}
