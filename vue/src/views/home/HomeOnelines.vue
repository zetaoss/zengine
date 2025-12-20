<script setup lang="ts">
import type { Avatar } from '@common/components/avatar/avatar'
import AvatarUser from '@common/components/avatar/AvatarUser.vue'
import httpy from '@common/utils/httpy'
import { onMounted,ref } from 'vue'

import linkify from '@/utils/linkify'

interface Row {
  id: number
  avatar: Avatar
  created: string
  message: string
}

const rows = ref<Row[]>([])

const load = async () => {
  const [data, err] = await httpy.get<Row[]>('/api/onelines/recent')
  if (err) {
    console.error(err)
    return
  }

  rows.value = await Promise.all(
    data.map(async (r) => ({
      ...r,
      message: await linkify(r.message),
    })),
  )
}

onMounted(load)
</script>

<template>
  <div v-for="r in rows" :key="r.id" class="py-2">
    <AvatarUser :avatar="r.avatar" />
    <span class="ml-1" v-html="r.message" />
    <span class="z-muted2 ml-1 text-xs">{{ r.created.substring(0, 10) }}</span>
  </div>
</template>
