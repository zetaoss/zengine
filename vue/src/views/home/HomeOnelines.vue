<script setup lang="ts">
import { ref } from 'vue'
import AvatarUserLink from '@common/components/avatar/AvatarUserLink.vue'
import type UserAvatar from '@common/types/userAvatar'
import http from '@/utils/http'
import linkify from '@/utils/linkify'

interface Row {
  id: number
  userAvatar: UserAvatar
  created: string
  message: string
}

const rows = ref([] as Row[])

async function fetchData() {
  const resp: any = await http.get('/api/onelines/recent')
  resp.data.forEach(async (r: Row) => {
    r.message = await linkify(r.message)
    rows.value.push(r)
  })
}

fetchData()
</script>

<template>
  <div v-for="r in rows" :key="r.id" class="py-2">
    <span class="silver">
      <AvatarUserLink :user-avatar="r.userAvatar" />
    </span>
    <span class="message ml-1" v-html="r.message" />
    <span class="silver ml-1 text-xs">{{ r.created.substring(0, 10) }}</span>
  </div>
</template>

<style scoped>
span {
  @apply break-all;
}
</style>
