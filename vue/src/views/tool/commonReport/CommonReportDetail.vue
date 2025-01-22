<script setup lang="ts">
import { ref } from 'vue'
import { mdiCheckBold, mdiContentCopy } from '@mdi/js'
import { useDateFormat, useDebounceFn } from '@vueuse/core'
import { useRoute, useRouter } from 'vue-router'

import TheIcon from '@common/components/TheIcon.vue'
import TheTooltip from '@common/components/TheTooltip.vue'
import AvatarUserLink from '@common/components/avatar/AvatarUserLink.vue'
import TheStar from './TheStar.vue'

import useAuthStore from '@/stores/auth'
import copyToClipboard from '@/utils/clipboard'
import http from '@/utils/http'

import type { Row } from './types'
import { getRatio, getScore, getWikitextTable } from './utils'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const id = Number(route.params.id as string)
const row = ref({} as Row)
const tooltips = ref([false, false, false])
const table = ref(null)
const tooltipStyle = { boxShadow: 'none', fontSize: '0.7rem', padding: '0.1rem 0.5rem' }

const debouncedFn = useDebounceFn(() => {
  tooltips.value = [false, false, false]
}, 3000)

function showTooltip(idx: number) {
  tooltips.value = [false, false, false]
  tooltips.value[idx] = true
  debouncedFn()
}

function getURL() {
  return `${window.location.hostname}/tool/common-report/${id}`
}

function copyURL() {
  showTooltip(0)
  copyToClipboard(getURL())
}

function copyTableHTML() {
  showTooltip(1)
  if (!table.value) return
  const el = table.value as HTMLTableElement
  const cleanHTML = el.outerHTML.replace(/ (data-v-[^=]+|class|rel|target|style)="[^"]*"/g, '')
    .replace(/<table>/g, '<table>\n')
    .replace(/<\/tr>/g, '</tr>\n')
    .replace(/<a [^>]*>(.*?)<\/a>/g, '$1')
  copyToClipboard(cleanHTML)
}

function copyTableWikitext() {
  showTooltip(2)
  if (!table.value) return
  const el = table.value as HTMLTableElement
  const s = getWikitextTable(el, id, getURL(), row.value.created_at)
  copyToClipboard(s)
}

async function del(r: Row) {
  if (!window.confirm(`${r.items[0].name} 등에 관한 #${r.id}번 통용보고서를 삭제하시겠습니까?`)) {
    return
  }
  await http.delete(`/api/common-report/${r.id}`)
  router.push({ path: '/tool/common-report' })
}

async function fetchData() {
  const resp = await http.get(`/api/common-report/${id}`)
  row.value = resp.data
}

fetchData()
</script>

<template>
  <div v-if="row.userAvatar" class="p-5">
    <div class="py-4 overflow-auto">
      <div class="float-right">
        <RouterLink to="/tool/common-report" class="btn">
          목록
        </RouterLink>
      </div>
    </div>
    <div class="border bg-z-card rounded p-5">
      <h2 class="my-5 text-2xl font-bold">
        통용 보고서 #{{ id }}
      </h2>
      <div class="h-10">
        <div class="float-left">
          <div>
            <AvatarUserLink :user-avatar="row.userAvatar" />
          </div>
          <div>
            <span class="text-slate-400">{{ useDateFormat(row.created_at, 'YYYY-MM-DD HH:mm').value }}</span>
          </div>
        </div>
        <div class="float-right">
          <TheTooltip :style="tooltipStyle" :show="tooltips[2]" content="Copied!">
            <button type="button" class="btn" @click="copyTableWikitext">
              표 복사(위키)
              <span v-if="!tooltips[2]">
                <TheIcon :path="mdiContentCopy" :size="14" />
              </span>
              <span v-else class="text-green-500">
                <TheIcon :path="mdiCheckBold" :size="14" />
              </span>
            </button>
          </TheTooltip>
        </div>
        <div class="float-right">
          <TheTooltip :style="tooltipStyle" :show="tooltips[1]" content="Copied!">
            <button type="button" class="btn" @click="copyTableHTML">
              표 복사(HTML)
              <span v-if="!tooltips[1]">
                <TheIcon :path="mdiContentCopy" :size="14" />
              </span>
              <span v-else class="text-green-500">
                <TheIcon :path="mdiCheckBold" :size="14" />
              </span>
            </button>
          </TheTooltip>
        </div>
        <div class="float-right">
          <TheTooltip :style="tooltipStyle" :show="tooltips[0]" content="Copied!">
            <button type="button" class="btn" @click="copyURL">
              URL 복사
              <span v-if="!tooltips[0]">
                <TheIcon :path="mdiContentCopy" :size="14" />
              </span>
              <span v-else class="text-green-500">
                <TheIcon :path="mdiCheckBold" :size="14" />
              </span>
            </button>
          </TheTooltip>
        </div>
      </div>
      <hr class="my-4 border-0 border-b">
      <div>
        <table ref="table" class="border-collapse">
          <tbody>
            <tr>
              <th colspan="2">표기</th>
              <td v-for="(item, idx) in row.items " :key="idx">
                <a class="new" :href="`/wiki/${item.name}`">{{ item.name }}</a>
              </td>
            </tr>
            <tr>
              <th colspan="2">판정</th>
              <td v-for="(_, idx) in row.items " :key="idx">
                <span v-if="idx == 0">
                  <TheStar :n="getScore(row)" />
                </span>
                <span v-else>—</span>
              </td>
            </tr>
            <tr>
              <th colspan="2">비율</th>
              <td v-for="(item, idx) in row.items " :key="idx">
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
              <td v-for="(item, idx) in row.items " :key="idx">
                {{ item.total.toLocaleString('en-US') }}
              </td>
            </tr>
            <tr>
              <th>다음</th>
              <th>블로그</th>
              <td v-for="(item, idx) in row.items " :key="idx">
                <a target="_blank" rel="noopener noreferrer" class="external"
                  :href="`http://search.daum.net/search?w=blog&q=${item.name}`">
                  {{ item.daum_blog.toLocaleString('en-US') }}
                </a>
              </td>
            </tr>
            <tr>
              <th rowspan="3">네이버</th>
              <th>블로그</th>
              <td v-for="(item, idx) in row.items " :key="idx">
                <a target="_blank" rel="noopener noreferrer" class="external"
                  :href="`https://search.naver.com/search.naver?where=post&query=${item.name}`">
                  {{ item.naver_blog.toLocaleString('en-US') }}
                </a>
              </td>
            </tr>
            <tr>
              <th>책</th>
              <td v-for="(item, idx) in row.items " :key="idx">
                <a target="_blank" rel="noopener noreferrer" class="external"
                  :href="`http://book.naver.com/search/search.nhn?query=${item.name}`">
                  {{ item.naver_book.toLocaleString('en-US') }}
                </a>
              </td>
            </tr>
            <tr>
              <th>뉴스</th>
              <td v-for="(item, idx) in row.items " :key="idx">
                <a target="_blank" rel="noopener noreferrer" class="external"
                  :href="`https://search.naver.com/search.naver?where=news&query=${item.name}`">
                  {{ item.naver_news.toLocaleString('en-US') }}
                </a>
              </td>
            </tr>
            <tr>
              <th colspan="2">구글</th>
              <td v-for="(item, idx) in row.items " :key="idx">
                <a target="_blank" rel="noopener noreferrer" class="external"
                  :href="`http://www.google.com/search?nfpr=1&q=%22${item.name}%22`">
                  {{ item.google_search.toLocaleString('en-US') }}
                </a>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <div class="py-4 overflow-auto">
      <div v-if="auth.canDelete(row.userAvatar.id)" class="float-left">
        <button type="button" class="btn" @click="del(row)">
          삭제
        </button>
      </div>
      <div v-if="auth.canWrite()" class="float-left">
        <button type="button" class="btn">
          재등록
        </button>
      </div>
      <div class="float-right">
        <RouterLink to="/tool/common-report" class="btn">
          목록
        </RouterLink>
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
