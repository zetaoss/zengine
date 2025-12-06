<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AvatarUser from '@common/components/avatar/AvatarUser.vue'
import type { Avatar } from '@common/components/avatar/avatar'
import http from '@/utils/http'
import linkify from '@/utils/linkify'

interface Row {
  id: number
  avatar: Avatar
  created: string
  message: string
}

const rows = ref<Row[]>([])

onMounted(async () => {
  try {
    const { data } = await http.get('/api/onelines/recent')
    rows.value = await Promise.all(
      data.map(async (r: Row) => ({
        ...r,
        message: await linkify(r.message),
      }))
    )
  } catch (error) {
    console.error('Error fetching onelines:', error)
  }
})
</script>

<template>
  <div v-for="r in rows" :key="r.id" class="py-2">
    <AvatarUser :avatar="r.avatar" />
    <span class="ml-1" v-html="r.message" />
    <span class="z-muted2 ml-1 text-xs">{{ r.created.substring(0, 10) }}</span>
  </div>
</template>
