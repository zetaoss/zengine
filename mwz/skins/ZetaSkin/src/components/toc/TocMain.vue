<script setup lang="ts">
import {
  type PropType, onMounted, ref, watch,
} from 'vue'

import { useThrottleFn, useWindowScroll } from '@vueuse/core'

import TocSection from './TocSection.vue'
import type { Section } from './types'

defineProps({
  datatoc: { type: {} as PropType<Section>, required: true },
})

const targetId = ref('')
const { y } = useWindowScroll()

onMounted(() => {
  const els = document.querySelectorAll('.mw-headline, #page-foot')
  const throttledFn = useThrottleFn(() => {
    let prev: Element | undefined
    let next: Element | undefined
    const topMargin = 50
    els.forEach((el) => {
      const { top } = el.getBoundingClientRect()
      if (top > topMargin && (next === undefined || top < next.getBoundingClientRect().top)) next = el
      if (top <= topMargin && (prev === undefined || top > prev.getBoundingClientRect().top)) prev = el
    })
    const closest = prev || next
    if (closest) targetId.value = closest.getAttribute('id') as string
  }, 200, true)
  watch(() => y.value, throttledFn)
})
</script>
<template>
  <div class="p-4 sticky top-0 text-sm">
    <aside class="overflow-y-auto z-scrollbar" :style="{ maxHeight: `calc(100vh - ${y > 48 ? 30 : 78 - y}px)` }">
      <ul class="list-none p-0 text-z-text2 border-l-2">
        <li class="px-4 m-0 font-bold">목차</li>
        <li class="m-0" v-for=" s in datatoc['array-sections'] " :key="s.index">
          <TocSection :targetId="targetId" :section="s" />
        </li>
      </ul>
    </aside>
  </div>
</template>
