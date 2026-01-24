<!-- TocApex.vue -->
<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import CapSticky from '@/components/CapSticky.vue'
import { useScrollSpy } from '@/composables/useScrollSpy'
import type { Section } from '@/types/toc'
import getRLCONF from '@/utils/rlconf'

import TocTree from './TocTree.vue'

const props = defineProps({
  headerOffset: { type: Number, default: 64 },
  side: { type: Boolean, default: false },
})

const { dataToc } = getRLCONF()

const tocObj = computed<Section | null>(() =>
  typeof dataToc === 'object' && dataToc !== null ? (dataToc as Section) : null
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
const tocRef = ref<HTMLElement | null>(null)
const isTocPast = ref(false)
const isDrawerMode = computed(() => !props.side && isTocPast.value)
const enableScrollSpy = computed(() => props.side || isDrawerMode.value)

const { activeIds: spyActiveIds } = useScrollSpy(anchors, props.headerOffset)
const activeIds = computed(() =>
  enableScrollSpy.value ? spyActiveIds.value : []
)

onMounted(() => {
  if (props.side || !('IntersectionObserver' in window)) return
  let observer: IntersectionObserver | null = null

  const observeToc = () => {
    if (observer) observer.disconnect()
    if (!tocRef.value) return
    observer = new IntersectionObserver(
      ([entry]) => {
        if (!entry) return
        isTocPast.value = entry.boundingClientRect.bottom <= props.headerOffset
      },
      { rootMargin: `-${props.headerOffset}px 0px 0px 0px`, threshold: 0 }
    )
    observer.observe(tocRef.value)
  }

  observeToc()
  watch(tocRef, observeToc)

  onBeforeUnmount(() => {
    if (observer) observer.disconnect()
  })
})

</script>

<template>
  <CapSticky v-if="side && tocObj" :marginY="16">
    <TocTree :toc="tocObj" :active-ids="activeIds" :header-offset="props.headerOffset" />
  </CapSticky>

  <div v-else-if="tocObj" ref="tocRef">
    <div class="inline-block rounded-md border border-slate-200 p-3 my-3">
      <TocTree :toc="tocObj" :active-ids="activeIds" :header-offset="props.headerOffset" :show-rail="false" />
    </div>
  </div>

</template>
