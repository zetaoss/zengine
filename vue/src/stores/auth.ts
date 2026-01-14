// @/stores/auth.ts
import { type UserInfo } from '@common/types/user'
import httpy from '@common/utils/httpy'
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

type MeResponse = { me: UserInfo | null }

const useAuthStore = defineStore('auth', () => {
  const userInfo = ref<UserInfo | null>(null)

  const isLoggedIn = computed(() => (userInfo.value?.id ?? 0) > 0)

  async function update() {
    const [res, err] = await httpy.get<MeResponse>('/api/me')
    if (err) {
      console.error(err)
      userInfo.value = null
      return
    }
    userInfo.value = res.me
  }

  function canWrite() {
    return isLoggedIn.value
  }

  function canEdit(id: number) {
    return (userInfo.value?.id ?? 0) === id
  }

  function canDelete(id: number) {
    const uid = userInfo.value?.id ?? 0
    const groups = userInfo.value?.groups ?? []
    return uid === id || groups.includes('sysop')
  }

  return { userInfo, update, isLoggedIn, canWrite, canEdit, canDelete }
})

export default useAuthStore
