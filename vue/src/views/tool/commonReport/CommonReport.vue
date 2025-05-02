<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import useAuthStore from '@/stores/auth'
import http from '@/utils/http'
import AvatarUserLink from '@common/components/avatar/AvatarUserLink.vue'
import TheSpinner from '@/components/TheSpinner.vue'
import TheStar from './TheStar.vue'
import ThePagination from '@/components/pagination/ThePagination.vue'
import type { PaginateData } from '@/components/pagination/types'
import CommonReportNew from './CommonReportNew.vue'
import type { Row } from './types'
import { getRatio, getScore } from './utils'
import { useRetrier } from './retrier'

interface RespData {
  current_page: number
  data: Row[]
  next_page_url: string | null
  path: string
  per_page: number
  prev_page_url: string | null
  to: number
  total: number
}

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const retrier = useRetrier(fetchData)

const reportData = ref<RespData | null>(null)
const paginateData = ref<PaginateData | null>(null)
const page = ref<number>(1)
const showModal = ref(false)

async function fetchData() {
  if (route.params.page) {
    page.value = Number(route.params.page)
  }
  try {
    const response = await http.get<RespData>('/api/common-report', {
      params: { page: String(page.value) }
    })
    reportData.value = response.data
    paginateData.value = { ...response.data, path: '/tool/common-report/page' } as PaginateData

    if (response.data.data.some(row => ['pending', 'running'].includes(row.phase))) {
      retrier.schedule()
    } else {
      retrier.clear()
    }
  } catch (error) {
    console.error('Error fetching common report data:', error)
  }
}

function openModal() {
  if (!auth.canWrite()) {
    router.push({ path: '/login', query: { redirect: '/tool/common-report' } })
    return
  }
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  retrier.start()
}

watch(
  () => route.params,
  () => {
    retrier.start()
  },
  { immediate: true }
)

onUnmounted(() => retrier.clear())
</script>

<template>
  <div class="p-5">
    <h2 class="my-5 text-2xl font-bold">통용 보고서</h2>
    <CommonReportNew :show="showModal" @close="closeModal" />
    <div v-if="reportData">
      <table class="w-full bg-z-card">
        <thead class="bg-z-head">
          <tr>
            <th class="not-mobile">
              번호
            </th>
            <th>이름</th>
            <th>건수</th>
            <th>비율</th>
            <th>판정</th>
          </tr>
        </thead>
        <tbody v-for="(row, rowKey) in reportData.data" :key="rowKey" class="align-top border-b border-[#88888866]">
          <tr v-for="(item, idx) in row.items" :key="idx">
            <td v-if="idx === 0" :rowspan="row.items.length" class="text-center text-sm">
              <RouterLink :to="`/tool/common-report/${row.id}`" class="btn dark:black w-full block border !p-1">
                #{{ row.id }} 상세보기
                <span v-if="row.phase === 'pending'">⏳</span>
                <span v-else-if="row.phase === 'running'" class="inline-block animate-spin">⏳</span>
                <span v-else-if="row.phase === 'failed'">❌</span>
              </RouterLink>
              <div>{{ row.created_at.substring(0, 10) }}</div>
              <div>
                <AvatarUserLink :user-avatar="row.userAvatar" />
              </div>
            </td>
            <td class="text-right">
              <a :href="`/wiki/${item.name}`">{{ item.name }}</a>
            </td>
            <td class="text-right">
              {{ item.total.toLocaleString('en-US') }}
            </td>
            <td v-if="getRatio(row, idx)">
              <div class="inline-block bg-[#77889966]" :style="{ width: (100 * getRatio(row, idx)) + '%' }">
                {{ (100 * getRatio(row, idx)).toFixed(1) }}%
              </div>
            </td>
            <td v-else>
              <br>
            </td>
            <td v-if="idx === 0">
              <TheStar :n="getScore(row)" />
            </td>
            <td v-else>
              —
            </td>
          </tr>
        </tbody>
      </table>
      <div class="py-4 overflow-auto">
        <div class="float-right">
          <button type="button" class="btn" :class="{ disabled: !auth.canWrite() }" @click="openModal">
            등록
          </button>
        </div>
      </div>
      <div class="text-center pb-8">
        <ThePagination v-if="paginateData" :paginate-data="paginateData" />
      </div>
    </div>
    <div v-else class="text-center">
      <TheSpinner size="2rem" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
th,
td {
  @apply p-2 px-1 md:px-2;
}
</style>
