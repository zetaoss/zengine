package statmodels

import "time"

var StatCFMetricNames = []string{
	"uniq_uniques",
	"sum_requests",
	"sum_bytes",
	"sum_cachedBytes",
	"sum_browserMap",
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
	Timeslot time.Time `gorm:"column:timeslot;primaryKey"`
	Name     string    `gorm:"column:name;primaryKey"`
	Value    string    `gorm:"column:value"`
}

type CFKV struct {
	Timeslot string `gorm:"column:timeslot;primaryKey"`
	Name     string `gorm:"column:name;primaryKey"`
	Value    string `gorm:"column:value"`
}
