<script setup lang="ts">
import { ref, watch } from 'vue'
import { useDateFormat } from '@vueuse/core'
import { useRoute } from 'vue-router'
import AvatarIcon from '@common/components/avatar/AvatarIcon.vue'
import ThePagination from '@/components/pagination/Pagination.vue'
import type { PaginateData } from '@/components/pagination/types'
import useAuthStore from '@/stores/auth'
import http from '@/utils/http'
import BoxPost from './box/BoxPost.vue'
import type { Post } from './types'
import ZSpinner from '@common/ui/ZSpinner.vue'
import RouterLinkButton from '@/ui/RouterLinkButton.vue'

const auth = useAuthStore()
const route = useRoute()

const postID = ref<number>(0)
const posts = ref<Post[]>([])
const paginateData = ref<PaginateData>({} as PaginateData)
const page = ref<number>(1)

const isLoading = ref(false)
const loadError = ref<string | null>(null)

const fetchData = async () => {
  postID.value = Number(route.params.id) || 0

  if (postID.value > 0) {
    isLoading.value = false
    posts.value = []
    return
  }

  page.value = Number(route.params.page) || 1
  isLoading.value = true
  loadError.value = null

  try {
    const { data } = await http.get(`/api/posts?page=${page.value}`)
    paginateData.value = { ...data, path: '/forum/page' }
    posts.value = data.data
  } catch (error) {
    console.error('failed to fetchData', error)
    loadError.value = '목록을 불러오지 못했습니다.'
  } finally {
    isLoading.value = false
  }
}

const formatDate = (date: string) => {
  return useDateFormat(date, 'YY-MM-DD HH:mm').value
}

watch(() => [route.params.id, route.params.page], fetchData, { immediate: true })
</script>

<template>
  <div class="p-5">
    <h2 class="my-5 text-2xl font-bold">포럼</h2>

    <!-- 상세 페이지 -->
    <div v-if="postID > 0">
      <div class="py-2">
        <div class="flex justify-end">
          <RouterLinkButton :to="{ path: `/forum/page/${page}` }">목록</RouterLinkButton>
        </div>
      </div>
      <BoxPost :post-i-d="postID" />
    </div>

    <!-- 목록 페이지 -->
    <div v-else class="text-sm">
      <div class="z-card">

        <!-- 항상 보이는 헤더 -->
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

        <!-- 로딩 -->
        <div v-if="isLoading" class="py-10 flex items-center justify-center text-gray-500">
          <ZSpinner />
        </div>

        <!-- 에러 -->
        <div v-else-if="loadError" class="py-10 text-center text-red-500">
          {{ loadError }}
        </div>

        <!-- 비어 있음 -->
        <div v-else-if="posts.length === 0" class="py-10 text-center text-gray-500">
          아직 등록된 글이 없습니다.
        </div>

        <!-- 정상 목록 -->
        <RouterLink v-else v-for="p in posts" :key="p.id" :to="{ path: `/forum/${p.id}` }"
          class="block md:flex py-2 px-3 md:px-2 border-b hover:no-underline z-text hover:bg-gray-50 dark:hover:bg-gray-800"
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
              <AvatarIcon :user-avatar="p.userAvatar" :size="15" />
              {{ p.userAvatar.name }}
            </span>
            <span class="w-auto md:w-[40%] md:text-center">
              {{ formatDate(p.created_at) }}
            </span>
            <span class="w-auto md:w-[15%] md:text-center">
              <span class="md:hidden px-1">· 조회</span>
              {{ p.hit }}
            </span>
          </div>
        </RouterLink>
      </div>

      <div v-if="!isLoading && !loadError" class="mt-4 text-right">
        <RouterLinkButton :to="{ path: `/forum/new` }" :disabled="!auth.canWrite()">글쓰기</RouterLinkButton>
      </div>

      <div v-if="!isLoading && !loadError" class="text-center py-4">
        <ThePagination :paginate-data="paginateData" />
      </div>
    </div>
  </div>
</template>
