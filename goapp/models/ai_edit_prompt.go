package models

import "time"

type AIEditPrompt struct {
	ID          int        `json:"id" gorm:"column:id"`
	UserID      int        `json:"user_id" gorm:"column:user_id"`
	UserName    string     `json:"user_name" gorm:"column:user_name"`
	Title       string     `json:"title" gorm:"column:title"`
	RequestType string     `json:"request_type" gorm:"column:request_type"`
	Content     string     `json:"content" gorm:"column:content"`
	Model       string     `json:"model" gorm:"column:model"`
	UseCount    int        `json:"use_count" gorm:"column:use_count"`
	UsedAt      *time.Time `json:"used_at" gorm:"column:used_at"`
	CreatedAt   string     `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   string     `json:"updated_at" gorm:"column:updated_at"`
}

func (AIEditPrompt) TableName() string {
	return "aiedit_prompts"
}
