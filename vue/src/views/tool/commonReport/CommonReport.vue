<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import useAuthStore from '@/stores/auth'
import http from '@/utils/http'
import UserAvatar from '@common/components/avatar/UserAvatar.vue'
import ZFold from '@common/ui/ZFold.vue'
import ZSpin from '@common/ui/ZSpin.vue'
import ZSpinner from '@common/ui/ZSpinner.vue'
import ZButton from '@common/ui/ZButton.vue'
import Star from './Star.vue'
import Pagination from '@/components/pagination/Pagination.vue'
import type { PaginateData } from '@/components/pagination/types'
import CommonReportNew from './CommonReportNew.vue'
import type { Row } from './types'
import RouterLinkButton from '@/ui/RouterLinkButton.vue'
import { getRatio, getScore } from './utils'
import { useRetrier } from './retrier'
import CProgressBar from '@common/components/CProgressBar.vue'

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
const loading = ref(false)

async function fetchData() {
  loading.value = true

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
  } finally {
    loading.value = false
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
  () => { retrier.start() },
  { immediate: true }
)

onUnmounted(() => retrier.clear())
</script>

<template>
  <div class="p-5">
    <h2 class="my-5 text-2xl font-bold">통용 보고서</h2>
    <CommonReportNew :show="showModal" @close="closeModal" />

    <table class="w-full">
      <thead class="z-table-header">
        <tr>
          <th class="not-mobile">번호</th>
          <th>이름</th>
          <th>건수</th>
          <th>비율</th>
          <th>판정</th>
        </tr>
      </thead>

      <thead v-if="loading">
        <tr>
          <th colspan="9" class="!p-0">
            <CProgressBar />
            <div v-if="!reportData" class="h-32 flex items-center justify-center">
              <ZSpinner />
            </div>
          </th>
        </tr>
      </thead>

      <template v-if="reportData">
        <tbody v-for="(row, rowKey) in reportData.data" :key="rowKey" class="align-top border-b border-[#88888866]">
          <tr v-for="(item, idx) in row.items" :key="idx">
            <td v-if="idx === 0" :rowspan="row.items.length" class="text-center text-sm">
              <RouterLinkButton class="w-full flex items-center justify-center gap-1"
                :to="`/tool/common-report/${row.id}`">
                <span>#{{ row.id }} 상세보기</span>
              </RouterLinkButton>
              <div>{{ row.created_at.substring(0, 10) }}</div>
              <div>
                <UserAvatar :user-avatar="row.userAvatar" />
              </div>
            </td>

            <td class="text-right">
              <a :href="`/wiki/${item.name}`">{{ item.name }}</a>
            </td>
            <td class="text-right">
              {{ item.total.toLocaleString('en-US') }}
            </td>
            <td>
              <div v-if="getRatio(row, idx)" class="inline-block bg-[#77889966]"
                :style="{ width: (100 * getRatio(row, idx)) + '%' }">
                {{ (100 * getRatio(row, idx)).toFixed(1) }}%
              </div>
            </td>
            <td>
              <template v-if="idx === 0">
                <span v-if="row.phase === 'pending'">
                  <ZFold>⏳</ZFold> Pending
                </span>
                <span v-else-if="row.phase === 'running'">
                  <ZSpin>⏳</ZSpin> Running
                </span>
                <span v-else-if="row.phase === 'failed'">
                  ❌ Error
                </span>
                <Star v-else :n="getScore(row)" />
              </template>
            </td>
          </tr>
        </tbody>
      </template>
    </table>

    <div class="py-2 flex justify-end">
      <ZButton :class="{ disabled: !auth.canWrite() }" @click="openModal">등록</ZButton>
    </div>

    <Pagination v-if="paginateData" :paginate-data="paginateData" />
  </div>
</template>

<style scoped>
table {

  th,
  td {
    padding: 0.5rem;
  }
}
</style>
