import type UserAvatar from '@common/types/userAvatar'

export interface Post {
  body: string
  cat: string
  created_at: string
  hit: number
  id: number
  is_notice: number
  replies_count: number
  tag_names: string[]
  tags_str: string
  title: string
  updated_at: string
  userAvatar: UserAvatar
  user_id: number
}

export interface Reply {
  body: string
  created_at: string
  id: number
  post_id: number
  userAvatar: UserAvatar
}

export interface DataError {
  title: string[]
  body: string[]
  errorType: string
  status: string
}

export interface Data {
  error: DataError
}

export interface ErrorResponse {
  data: Data
}
