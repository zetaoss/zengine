<!-- TocTree.vue -->
<script setup lang="ts">
import { scrollToTop } from '@common/utils/scroll'
import type { PropType } from 'vue'
import { computed } from 'vue'

import type { Section } from '@/types/toc'

import TocNode from './TocNode.vue'

defineOptions({ name: 'TocTree' })

const props = defineProps({
  toc: { type: Object as PropType<Section>, required: true },
  activeIds: { type: Array as PropType<string[]>, default: () => [] },
  headerOffset: { type: Number, default: 64 },
  showRail: { type: Boolean, default: true },
})

const sections = computed(() => props.toc['array-sections'] ?? [])

const scrollToAnchor = (id: string) => {
  const el = document.getElementById(id)
  if (!el) return
  const top = el.getBoundingClientRect().top + window.scrollY - props.headerOffset
  window.scrollTo({ top, behavior: 'smooth' })
}

const onNavigate = (id: string) => {
  if (!id) return
  scrollToAnchor(id)
  history.pushState(null, '', `#${id}`)
}
</script>

<template>
  <nav>
    <ul class="text-sm tracking-tight list-none m-0 p-0 z-muted">
      <li class="m-0 mb-2 flex items-center gap-1">
        <button type="button" @click="scrollToTop">
          목차
        </button>
      </li>
      <li v-for="s in sections" :key="s.index ?? s.anchor" class="m-0">
        <TocNode :section="s" :target-ids="activeIds" :depth="0" :show-rail="props.showRail"
          :on-navigate="onNavigate" />
      </li>
    </ul>
  </nav>
</template>
