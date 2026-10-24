package models

import "time"

// LLMCall is one model call of a chain step: the model zengine asked for, whether the step got a usable
// answer (a response that does not parse counts as a failure), and its cost. It is zengine's own record of
// which models work, kept across bob restarts.
type LLMCall struct {
	ID               uint64    `json:"id" gorm:"column:id;primaryKey"`
	Chain            string    `json:"chain" gorm:"column:chain"`
	ChainVersion     string    `json:"chain_version" gorm:"column:chain_version"`
	Step             string    `json:"step" gorm:"column:step"`
	Model            string    `json:"model" gorm:"column:model"`
	ResolvedModel    string    `json:"resolved_model" gorm:"column:resolved_model"`
	OK               bool      `json:"ok" gorm:"column:ok"`
	Error            string    `json:"error" gorm:"column:error"`
	PromptTokens     uint      `json:"prompt_tokens" gorm:"column:prompt_tokens"`
	CompletionTokens uint      `json:"completion_tokens" gorm:"column:completion_tokens"`
	ElapsedMs        uint      `json:"elapsed_ms" gorm:"column:elapsed_ms"`
	CreatedAt        time.Time `json:"created_at" gorm:"column:created_at"`
}

func (LLMCall) TableName() string {
	return "llm_calls"
}
