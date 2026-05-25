package statmodels

var StatGAMetricNames = []string{
	"active_users",
	"screen_page_views",
	"sessions",
}

type GA struct {
	Timeslot        string `gorm:"column:timeslot;primaryKey"`
	ActiveUsers     int    `gorm:"column:active_users"`
	ScreenPageViews int    `gorm:"column:screen_page_views"`
	Sessions        int    `gorm:"column:sessions"`
}
