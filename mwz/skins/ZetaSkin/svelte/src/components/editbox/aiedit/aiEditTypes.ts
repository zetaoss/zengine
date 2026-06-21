export type AIEditRequestType = 'create' | 'edit'

export type AIEditTaskPhase = 'Pending' | 'Generating' | 'Retrying' | 'Completed' | 'Rejected'

export interface AIEditPromptForRunner {
  id?: number
  userId: number
  title: string
  requestType: AIEditRequestType
  content: string
}

export interface AIEditRunnerSubmitPayload {
  requestType: AIEditRequestType
  title: string
  pageId?: number
  promptTitle: string
  llmInput: string
}

export interface AIEditStoreResult {
  ok: boolean
  id: number
  created: boolean
}

export interface AIEditTaskResult {
  id: number
  phase: AIEditTaskPhase
  llm_output: string | null
  last_error: string | null
  created_at: string
}
