<!-- ForumListPage.vue -->
<script setup lang="ts">
import './assets/forum-apex.css'

import useAuthStore from '@/stores/auth'
import RouterLinkButton from '@/ui/RouterLinkButton.vue'

import ForumPostList from './components/ForumPostList.vue'

const auth = useAuthStore()
</script>

<template>
  <div class="p-5">
    <h2 class="my-5 text-2xl font-bold">포럼</h2>

    <div class="flex justify-end mb-4">
      <RouterLinkButton to="/forum/new" :disabled="!auth.canWrite()">
        글쓰기
      </RouterLinkButton>
    </div>

    <ForumPostList />
  </div>
</template>
