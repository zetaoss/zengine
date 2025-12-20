<!-- TocApex.vue -->
<script setup lang="ts">
import type { PropType } from 'vue'
import { computed } from 'vue'
import type { Section } from './types'
import TocNode from './TocNode.vue'

import { scrollToTop, scrollToBottom } from '@common/utils/scroll'

import { useScrollSpy } from '@/composables/useScrollSpy'
import CapSticky from '@/components/CapSticky.vue'

const props = defineProps({
  toc: { type: [Object, String] as PropType<Section | string>, required: true },
  headerOffset: { type: Number, default: 64 },
})

const tocObj = computed<Section | null>(() =>
  typeof props.toc === 'object' && props.toc !== null ? (props.toc as Section) : null
)

const flattenAnchors = (root?: Section | null): string[] => {
  if (!root) return []
  const out: string[] = []
  const walk = (n: Section) => {
    if (n.anchor) out.push(n.anchor)
      ; (n['array-sections'] ?? []).forEach(walk)
  }
  walk(root)
  return out
}

const anchors = computed(() => flattenAnchors(tocObj.value))
const { activeIds } = useScrollSpy(anchors, props.headerOffset)

const scrollToAnchor = (id: string) => {
  const el = document.getElementById(id)
  if (!el) return
  const top = el.getBoundingClientRect().top + window.scrollY - props.headerOffset
  window.scrollTo({ top, behavior: 'smooth' })
}
</script>

<template>
  <CapSticky :marginY="16">
    <nav>
      <ul v-if="tocObj" class="text-sm tracking-tight list-none m-0 p-0 z-muted">
        <li class="m-0 mb-2 flex items-center gap-1">
          <button type="button" @click="scrollToTop">
            페이지 목차
          </button>
        </li>

        <li v-for="s in tocObj['array-sections'] ?? []" :key="s.index ?? s.anchor" class="m-0">
          <TocNode :section="s" :target-ids="activeIds" :depth="0" @navigate="scrollToAnchor" />
        </li>

        <li class="mt-2 opacity-50">
          <button @click="scrollToBottom">
            ∨
          </button>
        </li>
      </ul>
    </nav>
  </CapSticky>
</template>
