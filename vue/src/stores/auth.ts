import { computed, ref } from 'vue'

import { defineStore } from 'pinia'

import type UserData from '@common/types/userData'
import http from '@/utils/http'

const useAuthStore = defineStore('auth', () => {
  const userData = ref({} as UserData)

  const isLoggedIn = computed(() => userData.value.avatar && userData.value.avatar.id > 0)

  async function update() {
    const resp: any = await http.get('/api/me')
    userData.value = resp.data
  }

  function canWrite() {
    return isLoggedIn
  }

  function canEdit(id: number) {
    return userData.value.avatar && userData.value.avatar.id === id
  }

  function canDelete(id: number) {
    return userData.value.avatar.id === id || userData.value.groups.indexOf('sysop') !== -1
  }

  return {
    userData, update, isLoggedIn, canWrite, canEdit, canDelete,
  }
})

export default useAuthStore
