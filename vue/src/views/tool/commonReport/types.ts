import type UserAvatar from '@common/types/userAvatar'

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
  userAvatar: UserAvatar
  created_at: string
  updated_at: string
  state: number
  items: Item[]
}

export interface Score {
  star: number
  grade: string
}
