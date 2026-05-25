package models

type CommonReportItem struct {
	ID           int    `json:"id" gorm:"column:id"`
	ReportID     int    `json:"report_id" gorm:"column:report_id"`
	Name         string `json:"name" gorm:"column:name"`
	Total        int    `json:"total" gorm:"column:total"`
	DaumBlog     int    `json:"daum_blog" gorm:"column:daum_blog"`
	DaumBook     int    `json:"daum_book" gorm:"column:daum_book"`
	NaverBlog    int    `json:"naver_blog" gorm:"column:naver_blog"`
	NaverBook    int    `json:"naver_book" gorm:"column:naver_book"`
	NaverNews    int    `json:"naver_news" gorm:"column:naver_news"`
	GoogleSearch int    `json:"google_search" gorm:"column:google_search"`
}

func (CommonReportItem) TableName() string {
	return "zetawiki.common_report_item"
}
