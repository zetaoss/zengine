<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AvatarUserLink from '@common/components/avatar/AvatarUserLink.vue'
import type UserAvatar from '@common/types/userAvatar'
import http from '@/utils/http'
import linkify from '@/utils/linkify'

interface Row {
  page_title: string
  avatar: UserAvatar
  message: string
  page_name: string
}

const rows = ref<Row[]>([])

const fetchData = async () => {
  try {
    const { data } = await http.get('/api/comments/recent')
    rows.value = await Promise.all(
      data.map(async (x: Row) => ({
        ...x,
        message: await linkify(x.message),
      }))
    )
  } catch (error) {
    console.error('Failed to fetch comments:', error)
  }
}

onMounted(fetchData)
</script>

<template>
  <div v-for="r in rows" :key="r.page_title" class="py-2">
    <a :href="`/wiki/${r.page_title}`">{{ r.page_title.replace(/_/g, ' ') }}</a>
    <span class="silver ml-3">
      <AvatarUserLink :user-avatar="r.avatar" />
    </span>
    <div class="line-clamp-3 text-ellipsis break-words" v-html="r.message" />
  </div>
</template>
