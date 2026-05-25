package statmodels

var StatGSCMetricNames = []string{
	"clicks",
	"impressions",
	"ctr",
	"position",
}

type GSC struct {
	Timeslot    string  `gorm:"column:timeslot;primaryKey"`
	Clicks      int     `gorm:"column:clicks"`
	Impressions int     `gorm:"column:impressions"`
	Ctr         float64 `gorm:"column:ctr"`
	Position    float64 `gorm:"column:position"`
}
