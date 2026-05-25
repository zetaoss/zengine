package statmodels

import "time"

type StatNumeric struct {
	Timeslot        time.Time `gorm:"column:timeslot"`
	ActiveUsers     *float64  `gorm:"column:active_users"`
	ScreenPageViews *float64  `gorm:"column:screen_page_views"`
	Sessions        *float64  `gorm:"column:sessions"`
	Clicks          *float64  `gorm:"column:clicks"`
	Impressions     *float64  `gorm:"column:impressions"`
	Ctr             *float64  `gorm:"column:ctr"`
	Position        *float64  `gorm:"column:position"`
	Pages           *float64  `gorm:"column:pages"`
	Articles        *float64  `gorm:"column:articles"`
	Edits           *float64  `gorm:"column:edits"`
	Images          *float64  `gorm:"column:images"`
	Users           *float64  `gorm:"column:users"`
	ActiveUsersMW   *float64  `gorm:"column:activeusers"`
	Admins          *float64  `gorm:"column:admins"`
	Jobs            *float64  `gorm:"column:jobs"`
}
