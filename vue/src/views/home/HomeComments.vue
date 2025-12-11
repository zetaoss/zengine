<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AvatarUser from '@common/components/avatar/AvatarUser.vue'
import type { Avatar } from '@common/components/avatar/avatar'
import httpy from '@common/utils/httpy'
import linkify from '@/utils/linkify'

interface Row {
  page_title: string
  avatar: Avatar
  message: string
  page_name: string
}

const rows = ref<Row[]>([])

const load = async () => {
  const [data, err] = await httpy.get<Row[]>('/api/comments/recent')
  if (err) {
    console.error('recent comments', err)
    return
  }

  rows.value = await Promise.all(
    data.map(async (x) => ({
      ...x,
      message: await linkify(x.message),
    })),
  )
}

onMounted(load)
</script>

<template>
  <div v-for="r in rows" :key="r.page_title" class="py-2">
    <a :href="`/wiki/${r.page_title}`">{{ r.page_title.replace(/_/g, ' ') }}</a>
    <span class="silver ml-3">
      <AvatarUser :avatar="r.avatar" />
    </span>
    <div class="line-clamp-3 text-ellipsis break-words" v-html="r.message" />
  </div>
</template>
