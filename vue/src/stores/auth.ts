// auth.ts
import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import type UserData from '@common/types/userData'
import httpy from '@common/utils/httpy'

type MeResponse = {
  me: UserData | null
}

const useAuthStore = defineStore('auth', () => {
  const userData = ref<UserData | null>(null)

  const isLoggedIn = computed(() => {
    const a = userData.value?.avatar
    return !!a && a.id > 0
  })

  async function update() {
    const [res, err] = await httpy.get<MeResponse>('/api/me')
    if (err) {
      console.error(err)
      userData.value = null
      return
    }

    userData.value = res.me
  }

  function canWrite() {
    return isLoggedIn.value
  }

  function canEdit(id: number) {
    return userData.value?.avatar?.id === id
  }

  function canDelete(id: number) {
    const me = userData.value
    const uid = me?.avatar?.id ?? 0
    const groups = me?.groups ?? []
    return uid === id || groups.includes('sysop')
  }

  return {
    userData,
    update,
    isLoggedIn,
    canWrite,
    canEdit,
    canDelete,
  }
})

export default useAuthStore
