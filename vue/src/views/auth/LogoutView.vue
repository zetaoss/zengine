<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'

import httpy from '@common/utils/httpy'
import ZSpinner from '@common/ui/ZSpinner.vue'
import useAuthStore from '@/stores/auth'

const me = useAuthStore()
const router = useRouter()

onMounted(async () => {
  const [, err] = await httpy.get<unknown>('/api/logout')
  if (err) {
    console.error(err)
    return
  }

  await me.update()
  router.push({ path: '/login' })
})
</script>

<template>
  <div class="py-40 text-center">
    <ZSpinner />
    <div>Logging out...</div>
  </div>
</template>
