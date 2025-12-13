<!-- LogoutView.vue -->
<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute } from 'vue-router'

import httpy from '@common/utils/httpy'
import ZSpinner from '@common/ui/ZSpinner.vue'
import useAuthStore from '@/stores/auth'

const route = useRoute()
const me = useAuthStore()

onMounted(async () => {
  const [, err] = await httpy.get('/api/logout')
  if (err) {
    console.error(err)
    return
  }

  await me.update()

  const returnto = route.query.returnto as string | undefined

  if (returnto) {
    window.location.href = `/wiki/${returnto}`
  } else {
    window.location.href = '/login'
  }
})
</script>

<template>
  <div class="py-40 text-center">
    <ZSpinner />
    <div>Logging out...</div>
  </div>
</template>
