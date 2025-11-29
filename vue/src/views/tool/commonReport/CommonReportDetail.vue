<script setup lang="ts">
import { ref } from 'vue'
import { mdiCheckBold, mdiContentCopy } from '@mdi/js'
import { useDateFormat, useDebounceFn } from '@vueuse/core'
import { useRoute, useRouter } from 'vue-router'

import ZFold from '@common/ui/ZFold.vue'
import ZIcon from '@common/ui/ZIcon.vue'
import ZSpin from '@common/ui/ZSpin.vue'
import ZButton from '@common/ui/ZButton.vue'
import ZTooltip from '@common/ui/ZTooltip.vue'
import UserAvatar from '@common/components/avatar/UserAvatar.vue'
import { useToast } from '@common/composables/toast/useToast'
import { useConfirm } from '@common/composables/confirm/useConfirm'
import Star from './Star.vue'

import RouterLinkButton from '@/ui/RouterLinkButton.vue'
import useAuthStore from '@/stores/auth'
import copyToClipboard from '@/utils/clipboard'
import http from '@/utils/http'

import type { Row } from './types'
import { getRatio, getScore, getWikitextTable } from './utils'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const toast = useToast()
const confirm = useConfirm()

const id = Number(route.params.id as string)
const row = ref({} as Row)
const table = ref(null)

const tooltipStyle = { boxShadow: 'none', fontSize: '0.7rem', padding: '0.1rem 0.5rem' }

type ActionKey = 'wiki' | 'html' | 'url'

interface CopyAction {
  key: ActionKey
  label: string
  run: () => void
}

const activeTooltip = ref<ActionKey | null>(null)

const debouncedFn = useDebounceFn(() => {
  activeTooltip.value = null
}, 3000)

function getURL() {
  return `https://${window.location.hostname}/tool/common-report/${id}`
}

function copyURL() {
  copyToClipboard(getURL())
}

function copyTableHTML() {
  if (!table.value) return
  const el = table.value as HTMLTableElement
  const cleanHTML = el.outerHTML
    .replace(/ (data-v-[^=]+|class|rel|target|style)="[^"]*"/g, '')
    .replace(/<table>/g, '<table>\n')
    .replace(/<\/tr>/g, '</tr>\n')
    .replace(/<a [^>]*>(.*?)<\/a>/g, '$1')
  copyToClipboard(cleanHTML)
}

function copyTableWikitext() {
  if (!table.value) return
  const el = table.value
  const s = getWikitextTable(el, id, getURL(), row.value.created_at)
  copyToClipboard(s)
}

const copyActions: CopyAction[] = [
  { key: 'wiki', label: '표 복사 (WikiText)', run: copyTableWikitext },
  { key: 'html', label: '표 복사 (HTML)', run: copyTableHTML },
  { key: 'url', label: 'URL 복사', run: copyURL },
]

function handleCopy(action: CopyAction) {
  action.run()
  activeTooltip.value = action.key
  debouncedFn()
}

async function del(r: Row) {
  const ok = await confirm(`'${r.items[0].name}' 등에 관한 #${r.id}번 통용보고서를 삭제하시겠습니까?`)
  if (!ok) return

  await http.delete(`/api/common-report/${r.id}`)
  toast.show('삭제 완료')
  router.push({ path: '/tool/common-report' })
}

async function fetchData() {
  console.log('fetchData')
  const resp = await http.get(`/api/common-report/${id}`)
  row.value = resp.data
}

async function fetchDataWithRetry(retryDelay = 1000) {
  await fetchData()

  const phase = row.value.phase
  if (phase === 'pending' || phase === 'running') {
    setTimeout(() => {
      fetchDataWithRetry(retryDelay * 2)
    }, retryDelay)
  }
}

async function rerun(r: Row) {
  await http.post(`/api/common-report/${r.id}/rerun`)
  await fetchDataWithRetry()
}

fetchDataWithRetry()
</script>

<template>
  <div v-if="row.userAvatar" class="p-5">
    <div class="flex justify-end">
      <RouterLinkButton to="/tool/common-report">목록</RouterLinkButton>
    </div>
    <div class="border bg-z-card rounded p-5 my-2">
      <div class="my-5 flex items-center gap-3 text-2xl font-bold">
        <h2 class="m-0">통용 보고서 #{{ id }}</h2>
        <div class="flex items-center gap-1 text-base text-gray-600">
          <span v-if="row.phase === 'pending'">
            <ZFold>⏳</ZFold>
          </span>
          <span v-else-if="row.phase === 'running'">
            <ZSpin>⏳</ZSpin>
          </span>
          <span v-else-if="row.phase === 'failed'">❌</span>
          <span v-else-if="row.phase === 'succeeded'">✅</span>
          <span v-if="row.phase !== 'succeeded'">{{ row.phase }}</span>
        </div>
      </div>
      <div class="flex items-center">
        <div class="flex-1">
          <UserAvatar :user-avatar="row.userAvatar" />
          <div>{{ useDateFormat(row.created_at, 'YYYY-MM-DD HH:mm').value }}</div>
        </div>
        <div class="flex">
          <ZTooltip v-for="action in copyActions" :key="action.key" :style="tooltipStyle"
            :show="activeTooltip === action.key" content="Copied!">
            <ZButton @click="handleCopy(action)">
              <span class="mr-2">{{ action.label }}</span>
              <span v-if="activeTooltip !== action.key">
                <ZIcon :path="mdiContentCopy" />
              </span>
              <span v-else class="text-green-500">
                <ZIcon :path="mdiCheckBold" />
              </span>
            </ZButton>
          </ZTooltip>
        </div>
      </div>
      <hr class="my-4">
      <table ref="table" class="border-collapse">
        <tbody>
          <tr>
            <th colspan="2">표기</th>
            <td v-for="(item, idx) in row.items" :key="idx">
              <a class="new" :href="`/wiki/${item.name}`">{{ item.name }}</a>
            </td>
          </tr>
          <tr>
            <th colspan="2">판정</th>
            <td v-for="(_, idx) in row.items" :key="idx">
              <span v-if="idx == 0">
                <Star :n="getScore(row)" />
              </span>
              <span v-else>—</span>
            </td>
          </tr>
          <tr>
            <th colspan="2">비율</th>
            <td v-for="(item, idx) in row.items" :key="idx">
              <span v-if="getRatio(row, idx)">
                {{ (100 * getRatio(row, idx)).toFixed(1) }}%
              </span>
              <span v-else>
                —
              </span>
            </td>
          </tr>
          <tr>
            <th colspan="2">계</th>
            <td v-for="(item, idx) in row.items" :key="idx">
              {{ item.total.toLocaleString('en-US') }}
            </td>
          </tr>
          <tr>
            <th>다음</th>
            <th>블로그</th>
            <td v-for="(item, idx) in row.items" :key="idx">
              <a target="_blank" rel="noopener noreferrer" class="external"
                :href="`http://search.daum.net/search?w=blog&q=${item.name}`">
                {{ item.daum_blog.toLocaleString('en-US') }}
              </a>
            </td>
          </tr>
          <tr>
            <th rowspan="3">네이버</th>
            <th>블로그</th>
            <td v-for="(item, idx) in row.items" :key="idx">
              <a target="_blank" rel="noopener noreferrer" class="external"
                :href="`https://search.naver.com/search.naver?where=post&query=${item.name}`">
                {{ item.naver_blog.toLocaleString('en-US') }}
              </a>
            </td>
          </tr>
          <tr>
            <th>책</th>
            <td v-for="(item, idx) in row.items" :key="idx">
              <a target="_blank" rel="noopener noreferrer" class="external"
                :href="`http://book.naver.com/search/search.nhn?query=${item.name}`">
                {{ item.naver_book.toLocaleString('en-US') }}
              </a>
            </td>
          </tr>
          <tr>
            <th>뉴스</th>
            <td v-for="(item, idx) in row.items" :key="idx">
              <a target="_blank" rel="noopener noreferrer" class="external"
                :href="`https://search.naver.com/search.naver?where=news&query=${item.name}`">
                {{ item.naver_news.toLocaleString('en-US') }}
              </a>
            </td>
          </tr>
          <tr>
            <th colspan="2">구글</th>
            <td v-for="(item, idx) in row.items" :key="idx">
              <a target="_blank" rel="noopener noreferrer" class="external"
                :href="`http://www.google.com/search?nfpr=1&q=%22${item.name}%22`">
                {{ item.google_search.toLocaleString('en-US') }}
              </a>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div class="py-4 flex gap-2">
      <ZButton @click="del(row)" v-if="auth.canDelete(row.userAvatar.id)">
        삭제
      </ZButton>
      <ZButton @click="rerun(row)" v-if="auth.canDelete(row.userAvatar.id) && row.phase === 'failed'">
        재실행
      </ZButton>
      <div class="flex-1 text-right">
        <RouterLinkButton to="/tool/common-report">목록</RouterLinkButton>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
th,
td {
  @apply border text-right p-3;
}
</style>
