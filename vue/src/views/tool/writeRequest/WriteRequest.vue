<script setup lang="ts">
import { ref, watch } from 'vue'

import { useRoute, useRouter } from 'vue-router'

import ProgressBar from '@common/components/ProgressBar.vue'
import AvatarUser from '@common/components/avatar/AvatarUser.vue'
import type UserAvatar from '@common/types/userAvatar'
import ThePagination from '@/components/pagination/ThePagination.vue'
import type { PaginateData } from '@/components/pagination/types'
import useAuthStore from '@/stores/auth'
import http from '@/utils/http'

import WriteRequestNew from './WriteRequestNew.vue'

interface Row {
  hit: number
  id: number
  user_id: number
  rate: number
  ref: number
  title: string
  userAvatar: UserAvatar
  writed_at: string
  updated_at: string
}

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

interface Count {
  done: number
  todo: number
}

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const mode = ref('todo')
const respData = ref({} as RespData)
const paginateData = ref({} as PaginateData)
const page = ref(1)
const showModal = ref(false)
const count = ref({} as Count)
const loading = ref(true)

const retries = 0

async function fetchCount() {
  const resp = await http.get('/api/write-request/count')
  count.value = resp.data
}

async function fetchPage() {
  const resp = await http.get(`/api/write-request/${mode.value}`, { params: { page: String(page.value) } })
  respData.value = resp.data
  paginateData.value = resp.data
  paginateData.value.path = '/tool/write-request/page'
  loading.value = false
}

function fetchData() {
  loading.value = true
  fetchCount()
  if (route.params.page) {
    page.value = Number(String(route.params.page))
    if (retries < 1) window.scrollTo(0, 0)
  }
  fetchPage()
}

function openModal() {
  if (!auth.canWrite()) {
    router.push({ path: '/login', query: { redirect: '/tool/write-request' } })
    return
  }
  showModal.value = true
}
function closeModal() {
  showModal.value = false
  fetchData()
}

function setMode(m: string) {
  mode.value = m
  fetchData()
}

async function del(row: Row) {
  if (!window.confirm(`'${row.title}' 작성요청을 삭제하시겠습니까?`)) {
    return
  }
  try {
    await http.delete(`/api/write-request/${row.id}`)
    fetchData()
  } catch (err) {
    console.error(err)
  }
}

watch(() => route.params, fetchData)
fetchData()
</script>

<template>
  <div class="p-5">
    <h2 class="my-5 text-2xl font-bold">
      작성 요청
    </h2>
    <WriteRequestNew :show="showModal" @close="closeModal" />
    <div v-if="count" class="pb-3">
      <button type="button" class="inline-block p-3 border rounded-l"
        :class="[mode == 'todo' ? 'bg-slate-100 dark:bg-slate-700' : 'bg-white dark:bg-slate-900']"
        @click="setMode('todo')">
        요청 <span class="text-xs rounded-full px-1 bg-gray-400 text-white">{{ count.todo }}</span>
      </button>
      <button type="button" class="inline-block p-3 border border-l-0"
        :class="[mode == 'todo-top' ? 'bg-slate-100 dark:bg-slate-700' : 'bg-white dark:bg-slate-900']"
        @click="setMode('todo-top')">
        추천
      </button>
      <button type="button" class="inline-block p-3 border border-l-0 rounded-r"
        :class="[mode == 'done' ? 'bg-slate-100 dark:bg-slate-700' : 'bg-white dark:bg-slate-900']"
        @click="setMode('done')">
        완료 <span class="text-xs rounded-full px-2 bg-gray-400 text-white">{{ count.done }}</span>
      </button>
    </div>
    <table class="mytable w-full bg-z-card">
      <thead class="bg-z-head">
        <tr>
          <th>번호</th>
          <th>제목</th>
          <th>추천</th>
          <th>검색</th>
          <th>역링크</th>
          <th>요청일</th>
          <th>요청자</th>
        </tr>
      </thead>
      <thead>
        <tr>
          <th v-if="loading" colspan="9" class="!p-0">
            <ProgressBar />
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(row, rowKey) in respData.data" :key="rowKey" class="align-top border-b border-[#88888866]">
          <td class="text-center px-2">
            {{ row.id }}
          </td>
          <td class="w-[35%]">
            <a v-if="mode == 'todo'" :href="`/w/index.php?search=${row.title}`" class="new">{{ row.title }}</a>
            <a v-else-if="mode == 'todo-top'" :href="`/w/index.php?search=${row.title}`" class="new">{{ row.title }}</a>
            <a v-else :href="`/wiki/${row.title}`">{{ row.title }}</a>
            <button v-if="auth.canDelete(row.user_id)" type="button" class="btn-xs !px-1 ml-2 bg-orange-400 text-white"
              @click="del(row)">
              삭제
            </button>
          </td>
          <td class="text-center">
            {{ row.rate }}
          </td>
          <td class="text-center">
            {{ row.hit }}
          </td>
          <td class="text-center">
            <a :href="`/wiki/특수:가리키는문서/${row.title}`" class="btn">
              {{ row.ref }}
            </a>
          </td>
          <td class="text-center">
            {{ row.updated_at.substring(0, 10) }}
          </td>
          <td class="user">
            <AvatarUser :user-avatar="row.userAvatar" />
          </td>
        </tr>
      </tbody>
    </table>
    <div class=" py-4 overflow-auto">
      <div class="float-right">
        <button type="button" class="btn" :class="{ disabled: !auth.canWrite() }" @click="openModal">
          등록
        </button>
      </div>
    </div>
    <div class="text-center py-4 bg-slate-100 dark:bg-slate-900">
      <ThePagination :paginate-data="paginateData" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.mytable {

  th,
  td {
    @apply p-2 px-1 md:px-2;
  }
}
</style>
