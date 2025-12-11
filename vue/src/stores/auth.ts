// auth.ts
import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import type UserData from '@common/types/userData'

import httpy from '@common/utils/httpy'

const useAuthStore = defineStore('auth', () => {
  const userData = ref({} as UserData)

  const isLoggedIn = computed(
    () => !!userData.value.avatar && userData.value.avatar.id > 0,
  )

  async function update() {
    const [data, err] = await httpy.get<UserData>('/api/me')
    if (err) {
      console.error(err)
      return
    }
    userData.value = data
  }

  function canWrite() {
    return isLoggedIn
  }

  function canEdit(id: number) {
    return !!userData.value.avatar && userData.value.avatar.id === id
  }

  function canDelete(id: number) {
    return (
      !!userData.value.avatar &&
      (userData.value.avatar.id === id ||
        userData.value.groups.indexOf('sysop') !== -1)
    )
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
