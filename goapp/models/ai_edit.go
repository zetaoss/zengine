package models

import "time"

type AIEditPhase string

const (
	AIEditPhasePending    AIEditPhase = "Pending"
	AIEditPhaseGenerating AIEditPhase = "Generating"
	AIEditPhaseRetrying   AIEditPhase = "Retrying"
	AIEditPhaseCompleted  AIEditPhase = "Completed"
	AIEditPhaseRejected   AIEditPhase = "Rejected"
)

type AIEdit struct {
	ID           int         `json:"id" gorm:"column:id"`
	UserID       int         `json:"user_id" gorm:"column:user_id"`
	UserName     string      `json:"user_name" gorm:"column:user_name"`
	PageID       uint        `json:"page_id" gorm:"column:page_id"`
	PageTitle    string      `json:"page_title" gorm:"column:page_title"`
	RequestType  string      `json:"request_type" gorm:"column:request_type"`
	Phase        AIEditPhase `json:"phase" gorm:"column:phase"`
	LLMInput     string      `json:"llm_input" gorm:"column:llm_input"`
	LLMOutput    string      `json:"llm_output" gorm:"column:llm_output"`
	LLMModel     string      `json:"llm_model" gorm:"column:llm_model"`
	Revid        int         `json:"revid" gorm:"column:revid"`
	Attempts     int         `json:"attempts" gorm:"column:attempts"`
	ErrorCount   int         `json:"error_count" gorm:"column:error_count"`
	LastError    string      `json:"last_error" gorm:"column:last_error"`
	CreatedAt    time.Time   `json:"created_at" gorm:"column:created_at"`
	UpdatedAt    time.Time   `json:"updated_at" gorm:"column:updated_at"`
	RetryAt      *string     `json:"retry_at" gorm:"-"`
}

func (AIEdit) TableName() string {
	return "aiedit_tasks"
}
