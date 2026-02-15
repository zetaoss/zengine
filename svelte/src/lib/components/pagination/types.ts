export type PaginatePath = '/forum' | '/onelines' | '/tool/common-report' | '/tool/write-request'

export interface PaginateData {
  current_page: number
  last_page: number
  path: PaginatePath
}
