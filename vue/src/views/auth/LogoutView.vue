<script setup lang="ts">
import { onMounted } from 'vue'

import { useRouter } from 'vue-router'

import http from '@/utils/http'
import TheSpinner from '@common/ui/Spinner.vue'
import useAuthStore from '@/stores/auth'

const me = useAuthStore()
const router = useRouter()

onMounted(() => {
  http.get('/api/logout').then(() => {
    me.update()
    router.push({ path: '/login' })
  })
})
</script>

<template>
  <div class="py-40 text-center">
    <TheSpinner />
    <div>Logging out...</div>
  </div>
</template>
