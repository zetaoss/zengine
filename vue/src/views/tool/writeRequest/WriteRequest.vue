<script setup lang="ts">
import type { Avatar } from '@common/components/avatar/avatar'
import AvatarUser from '@common/components/avatar/AvatarUser.vue'
import CProgressBar from '@common/components/CProgressBar.vue'
import { useConfirm } from '@common/composables/confirm/useConfirm'
import { useToast } from '@common/composables/toast/useToast'
import ZButton from '@common/ui/ZButton.vue'
import ZIcon from '@common/ui/ZIcon.vue'
import ZSpinner from '@common/ui/ZSpinner.vue'
import httpy from '@common/utils/httpy'
import { mdiDelete } from '@mdi/js'
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import Pagination from '@/components/pagination/Pagination.vue'
import type { PaginateData } from '@/components/pagination/types'
import useAuthStore from '@/stores/auth'

import WriteRequestNew from './WriteRequestNew.vue'

interface Row {
  hit: number
  id: number
  user_id: number
  rate: number
  ref: number
  title: string
  avatar: Avatar
  writed_at: string
  updated_at: string
}

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

interface Count {
  done: number
  todo: number
}

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const toast = useToast()
const confirm = useConfirm()

const mode = ref<'todo' | 'todo-top' | 'done'>('todo')
const respData = ref({} as RespData)
const paginateData = ref({} as PaginateData)
const page = ref(1)
const showModal = ref(false)
const count = ref({} as Count)
const loading = ref(true)

const retries = 0

async function fetchCount() {
  const [data, err] = await httpy.get<Count>('/api/write-request/count')
  if (err) {
    console.error(err)
    return
  }

  count.value = data
}

async function fetchPage() {
  const [data, err] = await httpy.get<RespData>(
    `/api/write-request/${mode.value}`,
    { page: String(page.value) },
  )
  if (err) {
    console.error(err)
    loading.value = false
    return
  }

  respData.value = data
  paginateData.value = {
    ...(data as PaginateData),
    path: '/tool/write-request/page',
  }

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

function setMode(m: 'todo' | 'todo-top' | 'done') {
  mode.value = m
  fetchData()
}

async function del(row: Row) {
  const ok = await confirm(`'${row.title}' 작성요청을 삭제하시겠습니까 ? `)
  if (!ok) return

  const [, err] = await httpy.delete(`/api/write-request/${row.id}`)
  if (err) {
    console.error(err)
    return
  }

  fetchData()
  toast.show('삭제 완료')
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
    <table class="mytable w-full z-card">
      <thead class="z-table-header">
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
      <thead v-if="loading">
        <tr>
          <th colspan="9" class="!p-0">
            <CProgressBar />
            <div v-if="!respData.data" class="h-32 flex items-center justify-center">
              <ZSpinner />
            </div>
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
            <ZButton v-if="auth.canDelete(row.user_id)" color="ghost" class="text-[#888] py-1 align-middle leading-none"
              @click="del(row)">
              <ZIcon :path="mdiDelete" />
            </ZButton>
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
            <AvatarUser :avatar="row.avatar" />
          </td>
        </tr>
      </tbody>
    </table>
    <div class="py-4 text-right">
      <ZButton :disabled="!auth.canWrite()" @click="openModal">등록</ZButton>
    </div>
    <Pagination class="pb-4" :paginate-data="paginateData" />
  </div>
</template>

<style scoped>
th,
td {
  padding: .5rem 1rem;
}
</style>
