// laravel.ts
import type { Avatar } from '@common/components/avatar/avatar'
import { laravelApiGet, type ApiResult } from './util'

export interface UserInfo {
  user_id: number
  user_name: string
  user_registration: string
  user_editcount: number
  avatar: Avatar
}

export function getUserInfo(username: string): Promise<ApiResult<UserInfo>> {
  return laravelApiGet<UserInfo>(`/api/user/${encodeURIComponent(username)}`)
}

export type Stats = Record<string, number>

export function getStats(userId: number): Promise<ApiResult<Stats>> {
  return laravelApiGet(`/api/user/${userId}/stats`)
}
