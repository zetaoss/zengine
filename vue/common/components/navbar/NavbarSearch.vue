<script setup lang="ts">
import { ref } from 'vue'

import { mdiHistory, mdiMagnify, mdiShuffle } from '@mdi/js'
import { vOnClickOutside } from '@vueuse/components'

import ProgressBar from '../ProgressBar.vue'
import TheIcon from '../TheIcon.vue'
import CSSTooltip from '../tooltip/CSSTooltip.vue'

interface Page {
  description: string
  excerpt: string
  id: number
  key: string
  matched_title: string
  thumbnail: object
  title: string
}

const expanded = ref(false)
const searching = ref(false)
const keyword = ref('')
const query = ref('')
const pages = ref([] as Page[])
const index = ref(-1)

let timeoutID: ReturnType<typeof setTimeout>

function handleUpDown(offset: number) {
  clearTimeout(timeoutID)
  index.value += offset
  if (index.value < -1) index.value = pages.value.length
  if (index.value > pages.value.length) index.value = -1
  if (index.value > -1 && index.value < pages.value.length) query.value = pages.value[index.value].title
  else query.value = keyword.value
}

function close() {
  expanded.value = false
  index.value = -1
}

async function fetchData(q: string) {
  if (q.length < 1) return
  searching.value = true
  const response = await fetch(`/w/rest.php/v1/search/title?${new URLSearchParams({ q, limit: '10' })}`)
  const jsonData = await response.json()
  pages.value = jsonData.pages
  searching.value = false
  expanded.value = true
  index.value = -1
}

function goToSearch() {
  let search: string
  if (index.value < 0) search = query.value
  else if (index.value === pages.value.length) search = `${query.value}&fulltext=1&ns0=1`
  else search = encodeURIComponent(pages.value[index.value].title)
  close()
  window.location.href = `/w/index.php?title=특수:검색&search=${search}`
}

function onKeyUp() { handleUpDown(-1) }
function onKeyDown() { handleUpDown(1) }
function onKeyEnter() {
  setTimeout(() => {
    goToSearch()
  }, 500)
}

function onKeyEscape() {
  close()
}

function onInput(event: Event) {
  clearTimeout(timeoutID)
  timeoutID = setTimeout(() => {
    const target = event.target as HTMLInputElement
    const temp = target.value.trim()
    keyword.value = temp
    query.value = temp
    if (!temp) return
    fetchData(temp)
  }, 500)
}

function onBlur() {
  setTimeout(close, 500)
}
function onFocus() {
  if (pages.value.length < 0) return
  expanded.value = true
}
function onMouseover(i: number) {
  index.value = i
}
function onClick() {
  index.value = -1
  goToSearch()
}
function highlight(needle: string, haystack: string) {
  const re = new RegExp(needle.replace(/^\s*/, '').replace(/\s*$/, ''), 'gi')
  return haystack.replace(re, (s) => `<b>${s}</b>`)
}
</script>
<template>
  <div class="flex ml-auto md:max-w-2xl w-full">
    <div class="grow m-1">
      <div v-on-click-outside="close" class="relative rounded-t-lg bg-slate-200 dark:bg-gray-900"
        :class="{ 'rounded-b-lg': !expanded || keyword.length === 0 }" @keydown.up.prevent="onKeyUp"
        @keydown.down.prevent="onKeyDown" @keydown.enter="onKeyEnter" @keydown.escape.prevent="onKeyEscape">
        <div class="flex h-10">
          <input aria-label="search" type="search" class="grow px-3 h-full outline-0 bg-transparent" name="search"
            placeholder="Search..." title="검색 [alt-shift-f]" accesskey="f" autocomplete="off" :value="query"
            @input="onInput" @focus="onFocus" @blur="onBlur">
          <button type="button" class="flex-none w-12 h-full focus:z-10 focus:ring-1 bg-transparent focus:text-blue-700"
            @click="onClick">
            <TheIcon :path="mdiMagnify" />
          </button>
        </div>
        <div class="absolute z-10 w-full bg-slate-200 dark:bg-gray-900 rounded-b-lg border-t"
          :class="{ hidden: !expanded || keyword.length === 0 }">
          <ProgressBar :invisible="!searching" />
          <div v-if="pages">
            <div v-for="(p, i) in pages" :key="i">
              <a class="block p-2 px-3 text-z-text" :class="{ focused: index == i }"
                :href="`/w/index.php?title=특수:검색&search=${encodeURIComponent(p.title)}`" @mouseover="onMouseover(i)"
                @focus="onMouseover(i)" v-html="highlight(keyword, p.title)" />
            </div>
          </div>
          <a class="block p-2 px-3 border-t rounded-b-lg text-z-text"
            :href="`/w/index.php?title=특수:검색&fulltext=1&search=${keyword}`" :class="{ focused: index == pages.length }"
            @mouseover="onMouseover(pages.length)" @focus="onMouseover(pages.length)">
            <TheIcon :path="mdiMagnify" /> <b>{{ keyword }}</b> 항목이 포함된 글 검색
          </a>
        </div>
      </div>
    </div>
    <div class="flex-none">
      <CSSTooltip position="bottom" tooltipText="랜덤">
        <a href="/wiki/특수:임의문서" class="mybtn">
          <TheIcon :path="mdiShuffle" />
        </a>
      </CSSTooltip>
      <CSSTooltip position="bottom" tooltipText="최근바뀜">
        <a href="/wiki/특수:최근바뀜" class="mybtn">
          <TheIcon :path="mdiHistory" />
        </a>
      </CSSTooltip>
    </div>
  </div>
</template>
<style lang="scss" scoped>
.mybtn {
  @apply text-black dark:text-white no-underline block p-3 px-2;

  &:hover {
    @apply no-underline;
  }
}

.focused {
  @apply bg-[#07c] text-[#fff] no-underline;
}
</style>
