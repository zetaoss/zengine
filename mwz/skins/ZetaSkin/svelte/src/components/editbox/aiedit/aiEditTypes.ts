export type AIEditRequestType = 'create' | 'update'

export type AIEditTaskPhase = 'Pending' | 'Generating' | 'Retrying' | 'Completed' | 'Rejected'

export interface AIEditPromptForRunner {
  id?: number
  userId: number
  title: string
  requestType: AIEditRequestType
  content: string
  messages?: AIEditRenderedMessage[]
  model: string
}

// A user input of the chain a request type runs (doc-create, doc-update): a field of the run form.
export interface AIEditUserInput {
  name: string
  label?: string
  placeholder?: string
  type?: 'string' | 'text' | 'textList' | 'bool'
}

export type AIEditUserInputValue = string | boolean | string[]

export interface AIEditRequestPayload {
  requestType: AIEditRequestType
  title: string
  pageId?: number
  promptId?: number
  promptTitle: string
  wikitext: string
  userInputs: Record<string, AIEditUserInputValue>
}

export interface AIEditRenderedMessage {
  role: string
  content: string
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
