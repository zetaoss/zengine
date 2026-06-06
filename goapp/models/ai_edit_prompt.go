package models

type AIEditPrompt struct {
	ID          int    `json:"id" gorm:"column:id"`
	UserID      int    `json:"user_id" gorm:"column:user_id"`
	UserName    string `json:"user_name" gorm:"column:user_name"`
	Title       string `json:"title" gorm:"column:title"`
	RequestType string `json:"request_type" gorm:"column:request_type"`
	Content     string `json:"content" gorm:"column:content"`
	UseCount    int    `json:"use_count" gorm:"column:use_count"`
	CreatedAt   string `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   string `json:"updated_at" gorm:"column:updated_at"`
}

func (AIEditPrompt) TableName() string {
	return "aiedit_prompts"
}
