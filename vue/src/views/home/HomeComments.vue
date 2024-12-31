<script setup lang="ts">
import { ref } from 'vue'
import AvatarUserLink from '@common/components/avatar/AvatarUserLink.vue'
import type UserAvatar from '@common/types/userAvatar'
import linkify from '@/utils/linkify'

interface Row {
  page_title: string
  avatar: UserAvatar
  message: string
  page_name: string
}

const rows = ref([] as Row[])

async function fetchData() {
  const resp = await fetch('/api/comments/recent')
  const data = await resp.json() as Row[]
  rows.value = data
  rows.value.forEach((r, i) => {
    linkify(r.message).then((x) => { rows.value[i].message = x })
  })
}
fetchData()
</script>

<template>
  <div v-for="r in rows" :key="r.page_title" class="py-2">
    <a :href="'/wiki/' + r.page_title">{{ r.page_title.replace(/_/g, ' ') }}</a>
    <span class="silver ml-3">
      <AvatarUserLink :user-avatar="r.avatar" />
    </span>
    <div class="message" v-html="r.message" />
  </div>
</template>

<style lang="scss" scoped>
a.avatar {
  @apply text-gray-400;
}

.message {
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  display: -webkit-box;
  overflow: hidden;
  text-overflow: ellipsis;
  word-break: break-all;
}
</style>
