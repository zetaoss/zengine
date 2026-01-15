<!-- @/views/forum/components/ForumPostList.vue -->
<script setup lang="ts">
import AvatarIcon from '@common/components/avatar/AvatarIcon.vue'
import ZSpinner from '@common/ui/ZSpinner.vue'
import httpy from '@common/utils/httpy'
import { useDateFormat } from '@vueuse/core'
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import ThePagination from '@/components/pagination/ThePagination.vue'
import type { PaginateData } from '@/components/pagination/types'

import type { Post } from '../types'

defineProps<{
  currentPostId?: number
  title?: string
}>()

const route = useRoute()

const page = computed(() => {
  const p = Number(route.query.page)
  return Number.isFinite(p) && p > 0 ? p : 1
})

const posts = ref<Post[]>([])
const paginateData = ref<PaginateData | null>(null)
const isLoading = ref(false)
const loadError = ref<string | null>(null)

const formatDate = (date: string) =>
  useDateFormat(date, 'YY-MM-DD HH:mm').value

const fetchList = async () => {
  isLoading.value = true
  loadError.value = null

  const [data, err] = await httpy.get<{
    data: Post[]
    current_page: number
    last_page: number
  }>('/api/posts', { page: page.value })

  if (err) {
    loadError.value = '목록을 불러오지 못했습니다.'
    posts.value = []
    paginateData.value = null
    isLoading.value = false
    return
  }

  posts.value = data.data
  paginateData.value = {
    current_page: data.current_page,
    last_page: data.last_page,
    path: '/forum',
  }
  isLoading.value = false
}

watch(() => page.value, fetchList, { immediate: true })
</script>

<template>
  <div class="text-sm">
    <div v-if="title" class="mb-2 font-bold text-gray-700 dark:text-gray-300">
      {{ title }}
    </div>

    <div class="z-card">
      <div class="hidden md:flex p-2 font-bold text-center bg-[var(--z-table-header-bg)]">
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

      <div v-if="isLoading" class="py-10 flex items-center justify-center text-gray-500">
        <ZSpinner />
      </div>

      <div v-else-if="loadError" class="py-10 text-center text-red-500">
        {{ loadError }}
      </div>

      <div v-else-if="posts.length === 0" class="py-10 text-center text-gray-500">
        아직 등록된 글이 없습니다.
      </div>

      <RouterLink v-else v-for="p in posts" :key="p.id"
        :to="{ path: `/forum/${p.id}`, query: page === 1 ? {} : { page } }"
        class="block md:flex py-2 px-3 md:px-2 border-b hover:no-underline z-text hover:bg-gray-50 dark:hover:bg-gray-800"
        :class="{ 'bg-slate-100 dark:bg-stone-900': currentPostId === p.id }">
        <div class="flex py-1 md:w-[65%]">
          <span class="hidden md:inline w-[10%] text-center">{{ p.id }}</span>
          <span class="w-full md:w-[90%] pr-2 truncate">
            <span class="rounded-lg px-1.5 text-xs text-white dark:text-gray-200 bg-[#6668]">
              {{ p.cat }}
            </span>
            {{ p.title }}
            <span v-if="p.replies_count > 0">({{ p.replies_count }})</span>
          </span>
        </div>

        <div class="py-1 flex md:w-[35%]">
          <span class="flex-1 md:w-[45%] truncate">
            <AvatarIcon :user="{ id: p.user_id, name: p.user_name }" :size="15" />
            {{ p.user_name }}
          </span>
          <span class="md:w-[40%] md:text-center">
            {{ formatDate(p.created_at) }}
          </span>
          <span class="md:w-[15%] md:text-center">
            <span class="md:hidden px-1">· 조회</span>
            {{ p.hit }}
          </span>
        </div>
      </RouterLink>
    </div>

    <div v-if="!isLoading && !loadError" class="text-center py-4">
      <ThePagination v-if="paginateData" :paginate-data="paginateData" />
    </div>
  </div>
</template>
