<!-- TocMain.vue -->
<script setup lang="ts">
import type { PropType } from 'vue'
import { computed } from 'vue'
import type { Section } from './types'
import TocSection from './TocSection.vue'
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
  if (root) walk(root)
  return out
}

const anchors = computed(() => flattenAnchors(tocObj.value))
const { activeId } = useScrollSpy(anchors, props.headerOffset)

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
      <ul v-if="tocObj"
        class="text-sm text-z-text tracking-tight list-none m-0 p-0 pl-4 border-l-4 border-blue-400 dark:border-blue-600">
        <li class="opacity-50 m-0">페이지 목차</li>
        <li v-for="s in tocObj['array-sections'] ?? []" :key="s.index ?? s.anchor" class="m-0">
          <TocSection :section="s" :targetId="activeId ?? ''" @navigate="scrollToAnchor" />
        </li>
      </ul>
    </nav>
  </CapSticky>
</template>
