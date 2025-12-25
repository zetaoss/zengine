<!-- ForumViewPage.vue -->
<script setup lang="ts">
import './assets/forum-apex.css'

import { computed } from 'vue'
import { useRoute } from 'vue-router'

import RouterLinkButton from '@/ui/RouterLinkButton.vue'

import ForumPostList from './components/ForumPostList.vue'
import ViewerApex from './viewer/ViewerApex.vue'

const route = useRoute()

const postId = computed(() => Number(route.params.id) || 0)

const page = computed(() => {
  const p = Number(route.query.page)
  return Number.isFinite(p) && p > 0 ? p : 1
})
</script>

<template>
  <div class="p-5">
    <h2 class="my-5 text-2xl font-bold">포럼</h2>

    <div class="py-2 flex justify-end">
      <RouterLinkButton :to="{ path: '/forum', query: page === 1 ? {} : { page } }">
        목록
      </RouterLinkButton>
    </div>

    <ViewerApex :postId="postId" />

    <div class="mt-10">
      <ForumPostList :current-post-id="postId" :title="page === 1 ? '글 목록 (1페이지)' : `글 목록 (page ${page})`" />
    </div>
  </div>
</template>
