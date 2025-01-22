<script setup lang="ts">
import { ref, watch } from 'vue'
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

interface RespData {
  current_page: number
  data: Row[]
  next_page_url: string
  path: string
  per_page: number
  prev_page_url: string
  to: number
  total: number
}

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const respData = ref({} as RespData)
const paginateData = ref({} as PaginateData)
const page = ref(1)
const showModal = ref(false)

async function fetchData() {
  if (route.params.page) {
    page.value = Number(route.params.page as string)
    // window.scrollTo(0, 0)
  }
  const resp = await http.get('/api/common-report', { params: { page: `${page.value}` } })
  respData.value = resp.data
  paginateData.value = resp.data
  paginateData.value.path = '/tool/common-report/page'
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
  fetchData()
}

let retries = 0

function ensureCondition(conditionFunc, repeatFunc, maxRetries: number) {
  const timeHandler = () => {
    if (!conditionFunc() && retries <= maxRetries) {
      const timeout = 1000 * 1.5 ** retries
      console.log('retries', retries, 'timeout', timeout)
      repeatFunc()
      window.setTimeout(timeHandler, timeout)
      retries++
    }
  }
  timeHandler()
}

function isLoaded() {
  if (!respData.value.data || !(respData.value.data.every((v) => v.state > 1))) {
    return false
  }
  retries = 0
  return true
}

function ensureLoaded() {
  if (isLoaded() || retries > 0) return
  retries++
  ensureCondition(isLoaded, fetchData, 99)
}

watch(() => route.params, fetchData)
watch(respData, ensureLoaded)

fetchData()
</script>

<template>
  <div class="p-5">
    <h2 class="my-5 text-2xl font-bold">
      통용 보고서
    </h2>
    <CommonReportNew :show="showModal" @close="closeModal" />
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
      <tbody v-for="(row, rowKey) in respData.data" :key="rowKey" class="align-top border-b border-[#88888866]">
        <tr v-for="(item, idx) in row.items" :key="idx">
          <td v-if="idx == 0" :rowspan="row.items.length" class="text-center text-sm">
            <RouterLink :to="`/tool/common-report/${row.id}`" class="btn dark:black w-full block border !p-1">
              #{{ row.id }} 상세보기
              <span v-if="row.state == 0">
                <TheSpinner size=".9rem" extra-class="fill-red-700" />
              </span>
              <span v-if="row.state == 1">
                <TheSpinner size=".9rem" />
              </span>
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
            <div class="inline-block; bg-[#77889966]" :style="'width:' + (100 * getRatio(row, idx)) + '%'">
              {{ (100 * getRatio(row, idx)).toFixed(1) }}%
            </div>
          </td>
          <td v-else>
            <br>
          </td>
          <td v-if="idx == 0">
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
      <ThePagination :paginate-data="paginateData" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
th,
td {
  @apply p-2 px-1 md:px-2;
}
</style>
