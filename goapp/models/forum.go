package models

type ForumPost struct {
	ID           int      `json:"id" gorm:"column:id"`
	UserID       int      `json:"user_id" gorm:"column:user_id"`
	UserName     string   `json:"user_name" gorm:"column:user_name"`
	Cat          string   `json:"cat" gorm:"column:cat"`
	Title        string   `json:"title" gorm:"column:title"`
	Body         string   `json:"body" gorm:"column:body"`
	Hit          int      `json:"hit" gorm:"column:hit"`
	IsNotice     int      `json:"is_notice" gorm:"column:is_notice"`
	RepliesCount int      `json:"replies_count" gorm:"column:replies_count"`
	TagsStr      string   `json:"tags_str" gorm:"column:tags_str"`
	TagNames     []string `json:"tag_names" gorm:"-"`
	CreatedAt    string   `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    string   `json:"updated_at" gorm:"column:updated_at"`
}

type ForumReply struct {
	ID        int    `json:"id" gorm:"column:id"`
	PostID    int    `json:"post_id" gorm:"column:post_id"`
	UserID    int    `json:"user_id" gorm:"column:user_id"`
	UserName  string `json:"user_name" gorm:"column:user_name"`
	Body      string `json:"body" gorm:"column:body"`
	CreatedAt string `json:"created_at" gorm:"column:created_at"`
}
