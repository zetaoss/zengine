export type PaginatePath =
  | '/forum'
  | '/onelines'
  | '/tool/common-report'
  | '/tool/ai-prompts'
  | '/tool/write-request'
  | '/tool/write-request/done'
  | '/tool/write-request/todo'
  | '/tool/write-request/todo-top'

export interface PaginateData {
  current_page: number
  last_page: number
  path: PaginatePath
}
