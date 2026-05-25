package statmodels

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

type MWDaily struct {
	Timeslot    string `gorm:"column:timeslot;primaryKey"`
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
	Timeslot    string `gorm:"column:timeslot;primaryKey"`
	Pages       int    `gorm:"column:pages"`
	Articles    int    `gorm:"column:articles"`
	Edits       int    `gorm:"column:edits"`
	Images      int    `gorm:"column:images"`
	Users       int    `gorm:"column:users"`
	Activeusers int    `gorm:"column:activeusers"`
	Admins      int    `gorm:"column:admins"`
	Jobs        int    `gorm:"column:jobs"`
}
