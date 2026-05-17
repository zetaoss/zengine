package models

type CommonReportPhase string

const (
	CommonReportPhasePending   CommonReportPhase = "Pending"
	CommonReportPhaseRunning   CommonReportPhase = "Running"
	CommonReportPhaseSucceeded CommonReportPhase = "Succeeded"
	CommonReportPhaseFailed    CommonReportPhase = "Failed"
)

type CommonReport struct {
	ID        int                `json:"id" gorm:"column:id"`
	UserID    int                `json:"user_id" gorm:"column:user_id"`
	UserName  string             `json:"user_name" gorm:"column:user_name"`
	CreatedAt string             `json:"created_at" gorm:"column:created_at"`
	UpdatedAt string             `json:"updated_at" gorm:"column:updated_at"`
	Phase     CommonReportPhase  `json:"phase" gorm:"column:phase"`
	Items     []CommonReportItem `json:"items" gorm:"-"`
	Total     int                `json:"total" gorm:"-"`
}

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
