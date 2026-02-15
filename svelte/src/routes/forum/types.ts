export interface Post {
  id: number
  user_id: number
  user_name: string

  cat: string
  title: string
  body: string

  hit: number
  is_notice: number
  replies_count: number

  tags_str: string
  tag_names: string[]

  created_at: string
  updated_at: string
}

export interface Reply {
  id: number
  post_id: number
  user_id: number
  user_name: string

  body: string
  created_at: string
}
