package models

type ForumReply struct {
	ID        int    `json:"id" gorm:"column:id"`
	PostID    int    `json:"post_id" gorm:"column:post_id"`
	UserID    int    `json:"user_id" gorm:"column:user_id"`
	UserName  string `json:"user_name" gorm:"column:user_name"`
	Body      string `json:"body" gorm:"column:body"`
	CreatedAt string `json:"created_at" gorm:"column:created_at"`
}

func (ForumReply) TableName() string {
	return "replies"
}
