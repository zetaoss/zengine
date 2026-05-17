package models

import "time"

type WriteRequest struct {
	ID         int        `json:"id" gorm:"column:id"`
	UserID     int        `json:"user_id" gorm:"column:user_id"`
	UserName   string     `json:"user_name" gorm:"column:user_name"`
	WriterID   int        `json:"writer_id" gorm:"column:writer_id"`
	WriterName string     `json:"writer_name" gorm:"column:writer_name"`
	Rate       int        `json:"rate" gorm:"column:rate"`
	Ref        int        `json:"ref" gorm:"column:ref"`
	Title      string     `json:"title" gorm:"column:title"`
	WritedAt   *time.Time `json:"writed_at" gorm:"column:writed_at"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"column:updated_at"`
	Hit        int        `json:"hit" gorm:"column:hit"`
}

type WriteRequestTarget struct {
	ID       uint
	Title    string
	WritedAt *time.Time
}
