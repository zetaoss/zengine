package models

type RunboxPhase string

const (
	RunboxPhasePending   RunboxPhase = "Pending"
	RunboxPhaseRunning   RunboxPhase = "Running"
	RunboxPhaseSucceeded RunboxPhase = "Succeeded"
	RunboxPhaseFailed    RunboxPhase = "Failed"
)

type Runbox struct {
	Hash    string      `json:"hash" gorm:"primaryKey;column:hash"`
	Phase   RunboxPhase `json:"phase" gorm:"column:phase"`
	UserID  int         `json:"user_id" gorm:"column:user_id"`
	PageID  int         `json:"page_id" gorm:"column:page_id"`
	Type    string      `json:"type" gorm:"column:type"`
	Outs    string      `json:"outs" gorm:"column:outs"`
	CPU     float64     `json:"cpu" gorm:"column:cpu"`
	Mem     float64     `json:"mem" gorm:"column:mem"`
	Time    float64     `json:"time" gorm:"column:time"`
	Created string      `json:"created_at" gorm:"column:created_at"`
	Updated string      `json:"updated_at" gorm:"column:updated_at"`
}
