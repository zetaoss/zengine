package models

type EditBotPhase string

const (
	EditBotPhasePending          EditBotPhase = "Pending"
	EditBotPhaseGenerating       EditBotPhase = "Generating"
	EditBotPhasePublishing       EditBotPhase = "Publishing"
	EditBotPhaseRetryingGenerate EditBotPhase = "RetryingGenerate"
	EditBotPhaseRetryingPublish  EditBotPhase = "RetryingPublish"
	EditBotPhaseCompleted        EditBotPhase = "Completed"
	EditBotPhaseFailed           EditBotPhase = "Failed"
	EditBotPhaseRejected         EditBotPhase = "Rejected"
)

type EditBot struct {
	ID          int          `json:"id" gorm:"column:id"`
	UserID      int          `json:"user_id" gorm:"column:user_id"`
	UserName    string       `json:"user_name" gorm:"column:user_name"`
	Title       string       `json:"title" gorm:"column:title"`
	RequestType string       `json:"request_type" gorm:"column:request_type"`
	Phase       EditBotPhase `json:"phase" gorm:"column:phase"`
	LLMInput    string       `json:"llm_input" gorm:"column:llm_input"`
	LLMOutput   string       `json:"llm_output" gorm:"column:llm_output"`
	LLMModel    string       `json:"llm_model" gorm:"column:llm_model"`
	Revid       int          `json:"revid" gorm:"column:revid"`
	Attempts    int          `json:"attempts" gorm:"column:attempts"`
	ErrorCount  int          `json:"error_count" gorm:"column:error_count"`
	LastError   string       `json:"last_error" gorm:"column:last_error"`
	CreatedAt   string       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   string       `json:"updated_at" gorm:"column:updated_at"`
	RetryAt     *string      `json:"retry_at" gorm:"-"`
}

func (EditBot) TableName() string {
	return "edit_tasks"
}
