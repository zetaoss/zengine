import type { Avatar } from '@common/components/avatar/avatar'

export interface Item {
  id: number
  name: string
  total: number
  daum_blog: number
  daum_book: number
  naver_blog: number
  naver_book: number
  naver_news: number
  google_search: number
}

export interface Row {
  id: number
  total: number
  avatar: Avatar
  created_at: string
  updated_at: string
  phase: string
  items: Item[]
}
