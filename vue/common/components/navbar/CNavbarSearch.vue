<script setup lang="ts">
import { ref, computed, onBeforeUnmount } from 'vue'
import { useDebounceFn } from '@vueuse/core'
import { vOnClickOutside } from '@vueuse/components'

import { mdiHistory, mdiMagnify, mdiShuffle } from '@mdi/js'
import CProgressBar from '../CProgressBar.vue'
import BaseIcon from '@common/ui/BaseIcon.vue'

interface Page {
  description: string
  excerpt: string
  id: number
  key: string
  matched_title: string
  thumbnail: Record<string, unknown>
  title: string
}
interface SearchResponse { pages?: Page[];[k: string]: unknown }

const expanded = ref(false)
const searching = ref(false)
const keyword = ref('')
const pages = ref<Page[]>([])

const kIndex = ref(-1)
const hIndex = ref(-1)

const aborter = ref<AbortController | null>(null)

const displayQuery = computed(() =>
  kIndex.value >= 0 && kIndex.value < pages.value.length
    ? pages.value[kIndex.value].title
    : keyword.value
)

const currentIndex = computed(() => (hIndex.value >= 0 ? hIndex.value : kIndex.value))

function close() {
  expanded.value = false
  kIndex.value = -1
  hIndex.value = -1
}

async function fetchData(q: string) {
  const trimmed = q.trim()
  if (!trimmed) return

  aborter.value?.abort()
  const controller = new AbortController()
  aborter.value = controller

  searching.value = true
  try {
    const url = `/w/rest.php/v1/search/title?${new URLSearchParams({ q: trimmed, limit: '10' })}`
    const res = await fetch(url, { signal: controller.signal })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const json = (await res.json()) as SearchResponse
    pages.value = Array.isArray(json.pages) ? json.pages : []
    expanded.value = true
    kIndex.value = -1
    hIndex.value = -1
  } catch (e: unknown) {
    if (e instanceof DOMException && e.name === 'AbortError') return
    console.error('[search]', e)
  } finally {
    if (aborter.value === controller) aborter.value = null
    searching.value = false
  }
}

const debouncedFetch = useDebounceFn((q: string) => fetchData(q), 400)

function onInput(e: Event) {
  const val = (e.target as HTMLInputElement).value
  keyword.value = val
  kIndex.value = -1
  hIndex.value = -1

  if (!val.trim()) {
    aborter.value?.abort()
    pages.value = []
    expanded.value = false
    searching.value = false
    return
  }
  debouncedFetch(val)
}

function onFocus() {
  if (keyword.value.trim()) expanded.value = true
}

function handleUpDown(offset: number) {
  const max = pages.value.length
  let next = kIndex.value + offset
  if (next < -1) next = max
  if (next > max) next = -1
  kIndex.value = next
}

function goToSearch() {
  const base = new URL('/w/index.php', window.location.origin)
  base.searchParams.set('title', '특수:검색')

  if (kIndex.value < 0) {
    base.searchParams.set('search', displayQuery.value)
  } else if (kIndex.value === pages.value.length) {
    base.searchParams.set('search', displayQuery.value)
    base.searchParams.set('fulltext', '1')
    base.searchParams.set('ns0', '1')
  } else {
    base.searchParams.set('search', pages.value[kIndex.value].title)
  }

  close()
  window.location.href = base.toString()
}

const onKeyUp = () => handleUpDown(-1)
const onKeyDown = () => handleUpDown(1)
const onKeyEnter = () => goToSearch()
const onKeyEscape = () => close()

const onClick = () => { kIndex.value = -1; goToSearch() }

function escapeHTML(s: string) {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}
function highlight(needle: string, haystack: string) {
  const safeNeedle = needle.trim().replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  if (!safeNeedle) return escapeHTML(haystack)
  const re = new RegExp(safeNeedle, 'giu')
  return escapeHTML(haystack).replace(re, s => `<b>${escapeHTML(s)}</b>`)
}

onBeforeUnmount(() => { aborter.value?.abort() })
</script>

<template>
  <div class="flex ml-auto md:max-w-2xl w-full text-black dark:text-white">
    <div class="grow m-1">
      <div v-on-click-outside="close" class="relative" @keydown.up.prevent="onKeyUp" @keydown.down.prevent="onKeyDown"
        @keydown.enter="onKeyEnter" @keydown.escape.prevent="onKeyEscape">
        <div class="flex h-10 bg-white dark:bg-zinc-800 border rounded-t-lg"
          :class="{ 'rounded-b-lg': !expanded || !keyword.trim().length }">
          <input aria-label="search" type="search" class="grow px-3 h-full outline-0 bg-transparent" name="search"
            placeholder="Search..." title="검색 [alt-shift-f]" accesskey="f" autocomplete="off" :value="displayQuery"
            @input="onInput" @focus="onFocus" />
          <button type="button" class="flex-none w-12 h-full focus:z-10 focus:ring-1 bg-transparent focus:text-blue-700"
            @click="onClick">
            <BaseIcon :path="mdiMagnify" :size="24" />
          </button>
        </div>

        <div class="absolute z-40 w-full bg-white dark:bg-zinc-800 border border-t-0 rounded-b-lg"
          :class="{ hidden: !expanded || !keyword.trim().length }" @mouseleave="hIndex = -1">
          <CProgressBar :invisible="!searching" />

          <div v-if="pages.length">
            <div v-for="(p, i) in pages" :key="p.id">
              <a class="block p-2 px-3 text-z-text" :class="{ focused: currentIndex === i }"
                :href="`/w/index.php?title=특수:검색&search=${encodeURIComponent(p.title)}`" @mouseenter="hIndex = i"
                @focus="hIndex = i">
                <span v-html="highlight(keyword, p.title)" />
              </a>
            </div>
          </div>

          <!-- fulltext 행 -->
          <a class="block p-2 px-3 border-t rounded-b-lg text-z-text"
            :href="`/w/index.php?title=특수:검색&fulltext=1&search=${encodeURIComponent(keyword)}`"
            :class="{ focused: currentIndex === pages.length }" @mouseenter="hIndex = pages.length"
            @focus="hIndex = pages.length">
            <BaseIcon :path="mdiMagnify" />
            <b>{{ keyword }}</b> 항목이 포함된 글 검색
          </a>
        </div>
      </div>
    </div>

    <div class="flex-none flex">
      <a href="/wiki/특수:임의문서" class="navlink" title="랜덤">
        <BaseIcon :path="mdiShuffle" class="w-5 h-5" />
        <span class="hidden xl:inline ml-1">랜덤</span>
      </a>
      <a href="/wiki/특수:최근바뀜" class="navlink" title="바뀐글">
        <BaseIcon :path="mdiHistory" class="w-5 h-5" />
        <span class="hidden xl:inline ml-1">바뀐글</span>
      </a>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.focused {
  @apply bg-[#07c] text-white no-underline;
}
</style>
