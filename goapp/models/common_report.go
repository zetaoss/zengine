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

func (CommonReport) TableName() string {
	return "common_reports"
}
