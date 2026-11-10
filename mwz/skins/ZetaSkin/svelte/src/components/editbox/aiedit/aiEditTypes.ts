export type AIEditRequestType = 'create' | 'update'

export type AIEditTaskPhase = 'Pending' | 'Generating' | 'Retrying' | 'Completed' | 'Rejected'

export interface AIEditPromptForRunner {
  id?: number
  userId: number
  title: string
  requestType: AIEditRequestType
  // spec is the template as YAML (prompts.PromptTemplate); error is what stops it from running.
  spec: string
  error?: string
  model: string
}

// A user input of the chain a request type runs (doc-create, doc-update): a field of the run form (prompts.Input).
// type is read by the server from json_schema and ui; is_required ones must be filled to run, default is the
// starting value, and a choice is picked from json_schema.enum.
export interface AIEditUserInput {
  name: string
  description?: string
  type: 'string' | 'text' | 'textList' | 'bool' | 'choice'
  json_schema?: { type: string; enum?: string[] }
  ui?: { variant?: string; placeholder?: string }
  is_required?: boolean
  default?: string | boolean
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
