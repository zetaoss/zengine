import type { Avatar } from '@common/components/avatar/avatar'

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
  avatar: Avatar
  user_id: number
}

export interface Reply {
  body: string
  created_at: string
  id: number
  post_id: number
  avatar: Avatar
}
