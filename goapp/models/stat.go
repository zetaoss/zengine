package models

import "time"

var StatCFMetricNames = []string{
	"uniq_uniques",
	"sum_requests",
	"sum_bytes",
	"sum_cachedBytes",
	"sum_browserMap",
}

var StatGAMetricNames = []string{
	"active_users",
	"screen_page_views",
	"sessions",
}

var StatGSCMetricNames = []string{
	"clicks",
	"impressions",
	"ctr",
	"position",
}

var StatMWMetricNames = []string{
	"pages",
	"articles",
	"edits",
	"images",
	"users",
	"activeusers",
	"admins",
	"jobs",
}

var WorkerCFMetricNames = []string{
	"uniq_uniques",
	"sum_requests",
	"sum_pageViews",
	"sum_bytes",
	"sum_cachedBytes",
	"sum_cachedRequests",
	"sum_encryptedBytes",
	"sum_encryptedRequests",
	"sum_threats",
	"sum_browserMap",
	"sum_contentTypeMap",
	"sum_clientSSLMap",
	"sum_countryMap",
	"sum_ipClassMap",
	"sum_responseStatusMap",
	"sum_threatPathingMap",
}

type StatCF struct {
	Timeslot time.Time `gorm:"column:timeslot"`
	Name     string    `gorm:"column:name"`
	Value    string    `gorm:"column:value"`
}

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

type CFKV struct {
	Timeslot string `gorm:"column:timeslot"`
	Name     string `gorm:"column:name"`
	Value    string `gorm:"column:value"`
}

type GA struct {
	Timeslot        string `gorm:"column:timeslot"`
	ActiveUsers     int    `gorm:"column:active_users"`
	ScreenPageViews int    `gorm:"column:screen_page_views"`
	Sessions        int    `gorm:"column:sessions"`
}

type GSC struct {
	Timeslot    string  `gorm:"column:timeslot"`
	Clicks      int     `gorm:"column:clicks"`
	Impressions int     `gorm:"column:impressions"`
	Ctr         float64 `gorm:"column:ctr"`
	Position    float64 `gorm:"column:position"`
}

type MWDaily struct {
	Timeslot    string `gorm:"column:timeslot"`
	Pages       int    `gorm:"column:pages"`
	Articles    int    `gorm:"column:articles"`
	Edits       int    `gorm:"column:edits"`
	Images      int    `gorm:"column:images"`
	Users       int    `gorm:"column:users"`
	Activeusers int    `gorm:"column:activeusers"`
	Admins      int    `gorm:"column:admins"`
	Jobs        int    `gorm:"column:jobs"`
}

type MWHourly struct {
	Timeslot    string `gorm:"column:timeslot"`
	Pages       int    `gorm:"column:pages"`
	Articles    int    `gorm:"column:articles"`
	Edits       int    `gorm:"column:edits"`
	Images      int    `gorm:"column:images"`
	Users       int    `gorm:"column:users"`
	Activeusers int    `gorm:"column:activeusers"`
	Admins      int    `gorm:"column:admins"`
	Jobs        int    `gorm:"column:jobs"`
}
