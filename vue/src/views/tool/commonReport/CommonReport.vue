<!-- @/views/tool/commonReport/CommonReport.vue -->
<script setup lang="ts">
import AvatarUser from '@common/components/avatar/AvatarUser.vue'
import CProgressBar from '@common/components/CProgressBar.vue'
import ZButton from '@common/ui/ZButton.vue'
import ZFold from '@common/ui/ZFold.vue'
import ZSpin from '@common/ui/ZSpin.vue'
import ZSpinner from '@common/ui/ZSpinner.vue'
import httpy from '@common/utils/httpy'
import { onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import ThePagination from '@/components/pagination/ThePagination.vue'
import type { PaginateData } from '@/components/pagination/types'
import useAuthStore from '@/stores/auth'
import RouterLinkButton from '@/ui/RouterLinkButton.vue'
import { titlesExist } from '@/utils/mediawiki'

import CommonReportNew from './CommonReportNew.vue'
import { useRetrier } from './retrier'
import TheStar from './TheStar.vue'
import type { Row } from './types'
import { getRatio, getScore } from './utils'

interface RespData {
  current_page: number
  data: Row[]
  last_page: number
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
const titleExists = ref<Record<string, boolean>>({})

function resolvePageFromRoute(): number {
  const queryPage = Number(String(route.query.page ?? ''))
  if (Number.isFinite(queryPage) && queryPage > 0) return queryPage

  const paramPage = Number(String(route.params.page ?? ''))
  if (Number.isFinite(paramPage) && paramPage > 0) return paramPage

  return 1
}

async function fetchData() {
  loading.value = true

  page.value = resolvePageFromRoute()

  const [data, err] = await httpy.get<RespData>('/api/common-report', {
    page: String(page.value),
  })

  if (err) {
    console.error('Error fetching common report data:', err)
    loading.value = false
    return
  }

  reportData.value = data
  void syncTitleExists()
  paginateData.value = {
    current_page: data.current_page,
    last_page: data.last_page,
    path: '/tool/common-report',
  }

  if (data.data.some(row => ['pending', 'running'].includes(row.phase))) {
    retrier.schedule()
  } else {
    retrier.clear()
  }

  loading.value = false
}

async function syncTitleExists() {
  const rows = reportData.value?.data ?? []
  const titles = rows.flatMap(row => row.items.map(item => item.name))
  if (titles.length === 0) return
  const existsMap = await titlesExist(titles)
  Object.assign(titleExists.value, existsMap)
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

function titleState(name: string): 'unknown' | '' | 'new' {
  const exists = titleExists.value[name]
  return exists === undefined ? 'unknown' : exists ? '' : 'new'
}

function wikiHref(name: string): string {
  return titleState(name) === 'new'
    ? `/wiki/${name}/edit?redlink=1`
    : `/wiki/${name}`
}

watch(
  () => [route.params.page, route.query.page],
  () => { retrier.start() },
  { immediate: true },
)

onUnmounted(() => retrier.clear())
</script>

<template>
  <div class="p-5">
    <h2 class="my-5 text-2xl font-bold">통용 보고서</h2>
    <CommonReportNew :show="showModal" @close="closeModal" />

    <table class="w-full">
      <thead class="z-base3">
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
          <th colspan="9" class="p-0!">
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
                <AvatarUser :user="{ id: row.user_id, name: row.user_name }" />
              </div>
            </td>

            <td class="text-right">
              <a :href="wikiHref(item.name)" :class="titleState(item.name)">
                {{ item.name }}
              </a>
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
                <TheStar v-else :n="getScore(row)" />
              </template>
            </td>
          </tr>
        </tbody>
      </template>
    </table>

    <div class="py-2 flex justify-end">
      <ZButton :class="{ disabled: !auth.canWrite() }" @click="openModal">등록</ZButton>
    </div>

    <ThePagination v-if="paginateData" :paginate-data="paginateData" />
  </div>
</template>

<style scoped>
table {

  th,
  td {
    padding: 0.5rem;
  }
}

.unknown {
  color: gray;
}
</style>
