<!-- @/views/misc/TheOnelines.vue -->
<script setup lang="ts">
import AvatarUser from '@common/components/avatar/AvatarUser.vue'
import { showConfirm } from '@common/ui/confirm/confirm'
import { showToast } from '@common/ui/toast/toast'
import ZButton from '@common/ui/ZButton.vue'
import ZIcon from '@common/ui/ZIcon.vue'
import ZSpinner from '@common/ui/ZSpinner.vue'
import httpy from '@common/utils/httpy'
import { mdiDelete } from '@mdi/js'
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

import ThePagination from '@/components/pagination/ThePagination.vue'
import type { PaginateData } from '@/components/pagination/types'
import useAuthStore from '@/stores/auth'
import linkify from '@/utils/linkify'

interface Row {
  id: number
  user_id: number
  user_name: string
  created: string
  message: string
}

const route = useRoute()
const auth = useAuthStore()

const page = computed(() => {
  const p = Number(route.query.page)
  return Number.isFinite(p) && p > 0 ? p : 1
})

const rows = ref<Row[]>([])
const paginateData = ref<PaginateData | null>(null)
const isLoading = ref(false)
const loadError = ref<string | null>(null)

const fetchList = async () => {
  isLoading.value = true
  loadError.value = null

  const [data, err] = await httpy.get<{
    data: Row[]
    current_page: number
    last_page: number
  }>('/api/onelines', { page: page.value })

  if (err) {
    console.error(err)
    rows.value = []
    paginateData.value = null
    loadError.value = '목록을 불러오지 못했습니다.'
    isLoading.value = false
    return
  }

  rows.value = await Promise.all(
    data.data.map(async (r) => ({
      ...r,
      message: await linkify(r.message),
    })),
  )
  paginateData.value = {
    current_page: data.current_page,
    last_page: data.last_page,
    path: '/onelines',
  }
  isLoading.value = false
}

const del = async (row: Row) => {
  if (!auth.canDelete(row.user_id)) return
  const ok = await showConfirm('이 한줄잡담을 삭제하시겠습니까 ? ')
  if (!ok) return

  const [, err] = await httpy.delete(`/api/onelines/${row.id}`)
  if (err) {
    console.error(err)
    showToast(err.message || '삭제 실패')
    return
  }

  showToast('삭제 완료')
  fetchList()
}

watch(() => page.value, fetchList, { immediate: true })
</script>

<template>
  <div class="p-5">
    <h2 class="my-5 text-2xl font-bold">한줄잡담</h2>

    <div class="z-card">
      <div v-if="isLoading" class="py-10 flex items-center justify-center text-gray-500">
        <ZSpinner />
      </div>

      <div v-else-if="loadError" class="py-10 text-center text-red-500">
        {{ loadError }}
      </div>

      <div v-else-if="rows.length === 0" class="py-10 text-center text-gray-500">
        아직 등록된 한줄잡담이 없습니다.
      </div>

      <div v-else>
        <div v-for="r in rows" :key="r.id" class="py-3 px-3 border-b last:border-b-0">
          <AvatarUser :user="{ id: r.user_id, name: r.user_name }" />
          <span class="ml-1" v-html="r.message" />
          <span class="z-muted2 ml-1 text-xs">{{ r.created.substring(0, 10) }}</span>
          <ZButton v-if="auth.canDelete(r.user_id)" color="ghost" class="text-[#888] py-1 align-middle leading-none"
            @click="del(r)">
            <ZIcon :path="mdiDelete" />
          </ZButton>
        </div>
      </div>
    </div>

    <div v-if="!isLoading && !loadError" class="text-center py-4">
      <ThePagination v-if="paginateData" :paginate-data="paginateData" />
    </div>
  </div>
</template>
