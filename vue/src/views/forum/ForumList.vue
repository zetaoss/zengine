<script setup lang="ts">
import { ref, watch } from 'vue'

import { useDateFormat } from '@vueuse/core'
import { useRoute } from 'vue-router'

import AvatarCore from '@common/components/avatar/AvatarCore.vue'
import ThePagination from '@/components/pagination/ThePagination.vue'
import type { PaginateData } from '@/components/pagination/types'
import useAuthStore from '@/stores/auth'
import http from '@/utils/http'

import BoxPost from './box/BoxPost.vue'
import type { Post } from './types'

const auth = useAuthStore()
const route = useRoute()

const postID = ref(0)
const posts = ref([] as Post[])
const paginateData = ref({} as PaginateData)
const page = ref(1)

async function fetchData() {
  if (route.params.id) {
    postID.value = Number(route.params.id as string)
    window.scrollTo(0, 80)
    return
  }
  postID.value = 0
  if (route.params.page) {
    page.value = Number(route.params.page as string)
  }
  const resp: any = await http.get(`/api/posts?page=${page.value}`)
  paginateData.value = resp.data
  paginateData.value.path = '/forum/page'
  posts.value = resp.data.data
}

watch(() => route.params, fetchData)
fetchData()
</script>

<template>
  <div class="p-5">
    <h2 class="my-5 text-2xl font-bold">
      포럼
    </h2>
    <div v-if="postID > 0">
      <div class="py-2 overflow-auto">
        <div class="float-right">
          <RouterLink :to="{ path: `/forum/page/${page}` }" class="btn">
            목록
          </RouterLink>
        </div>
      </div>
      <BoxPost :post-i-d="postID" />
    </div>
    <div v-if="posts.length > 0" class="text-sm">
      <div class="bg-z-card">
        <div class="hidden md:flex p-2 bg-z-head font-bold text-center">
          <div class="flex w-[65%]">
            <span class="w-[10%]">번호</span>
            <span class="w-[90%] text-left">제목</span>
          </div>
          <div class="flex w-[35%]">
            <span class="w-[45%] text-left">작성자</span>
            <span class="w-[40%]">작성일</span>
            <span class="w-[15%]">조회</span>
          </div>
        </div>
        <RouterLink v-for="p in posts" :key="p.id" :to="{ path: `/forum/${p.id}` }"
          class="block md:flex py-2 px-3 md:px-2 border-b hover:no-underline text-black dark:text-slate-300 hover:bg-gray-50 dark:hover:bg-gray-800"
          :class="{ 'bg-slate-100 dark:bg-stone-900': postID == p.id }">
          <div class="flex py-1 md:w-[65%]">
            <span class="hidden md:inline w-[10%] text-center">{{ p.id }}</span>
            <span class="w-full md:w-[90%] truncate">
              <span class="rounded-lg px-1.5 text-xs text-white dark:text-gray-200 bg-[#6668]">
                {{ p.cat }}
              </span>
              {{ p.title }}
              <span v-if="p.replies_count > 0">({{ p.replies_count }})</span>
            </span>
          </div>
          <div class="py-1 flex md:w-[35%]">
            <span class="flex-1 w-auto md:inline md:w-[45%] truncate">
              <AvatarCore :user-avatar="p.userAvatar" :size="15" />
              {{ p.userAvatar.name }}
            </span>
            <span class="w-auto md:w-[40%] md:text-center">
              {{ useDateFormat(p.created_at, 'YY-MM-DD HH:mm').value }}
            </span>
            <span class="w-auto md:w-[15%] md:text-center"><span class="md:hidden px-1">· 조회</span>
              {{ p.hit }}
            </span>
          </div>
        </RouterLink>
      </div>
      <div class="mt-4 text-right">
        <RouterLink :to="{ path: `/forum/new` }" class="btn" :class="{ disabled: !auth.canWrite() }">
          글쓰기
        </RouterLink>
      </div>
      <div class="text-center py-4">
        <ThePagination :paginate-data="paginateData" />
      </div>
    </div>
  </div>
</template>

<style>
#post-body p {
  margin-bottom: 1rem;
}

.cnt:before {
  content: '| 조회';
}
</style>
