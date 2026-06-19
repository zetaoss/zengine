export type AIEditRequestType = 'create' | 'edit'

export type AIEditRunnerLayout = 'full' | 'compact'

export type ExistingContentState =
  | 'idle'
  | 'loading'
  | 'ready'
  | 'stale'
  | 'not_found'
  | 'error'
  | 'not_required'
  | 'available'
  | 'invalid'

export interface AIEditPromptForRunner {
  id?: number
  title: string
  requestType: AIEditRequestType
  content: string
}

export interface AIEditExistingContentResult {
  title: string
  content: string
  pageId?: number
}

export interface AIEditRunnerSubmitPayload {
  requestType: AIEditRequestType
  title: string
  pageId?: number
  promptTitle: string
  llmInput: string
}
