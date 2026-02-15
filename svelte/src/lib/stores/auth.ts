// src/lib/stores/auth.ts
import { derived, writable } from 'svelte/store'

import type { UserInfo } from '$shared/types/user'
import httpy from '$shared/utils/httpy'

type MeResponse = { me: UserInfo | null }

const userInfo = writable<UserInfo | null>(null)

const isLoggedIn = derived(userInfo, ($userInfo) => ($userInfo?.id ?? 0) > 0)

async function update() {
  const [res, err] = await httpy.get<MeResponse>('/api/me')
  if (err) {
    console.error(err)
    userInfo.set(null)
    return
  }
  userInfo.set(res.me)
}

const canWrite = derived(isLoggedIn, ($isLoggedIn) => $isLoggedIn)

const canEdit = derived(userInfo, ($userInfo) => {
  return (id: number) => ($userInfo?.id ?? 0) === id
})

const canDelete = derived(userInfo, ($userInfo) => {
  return (id: number) => {
    const uid = $userInfo?.id ?? 0
    const groups = $userInfo?.groups ?? []
    return uid === id || groups.includes('sysop')
  }
})

const store = {
  userInfo,
  update,
  isLoggedIn,
  canWrite,
  canEdit,
  canDelete,
}

export default function useAuthStore() {
  return store
}
