export type PaginatePath = '/forum' | '/onelines' | '/tool/common-report' | '/tool/ai-edit' | '/tool/ai-edit/tasks' | '/tool/write-request'

export interface PaginateData {
  current_page: number
  last_page: number
  path: PaginatePath
}
