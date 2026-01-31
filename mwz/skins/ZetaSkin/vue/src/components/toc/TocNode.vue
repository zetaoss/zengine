<!-- TocNode.vue -->
<script setup lang="ts">
import type { PropType } from 'vue'
import { computed } from 'vue'

import type { Section } from '@/types/toc'

defineOptions({ name: 'TocNode' })

const props = defineProps({
  section: { type: Object as PropType<Section>, required: true },
  targetIds: { type: Array as PropType<string[]>, required: true },
  depth: { type: Number, default: 0 },
  showRail: { type: Boolean, default: true },
  onNavigate: {
    type: Function as PropType<(id: string) => void>,
    required: true,
  },
})

const anchor = computed(() => props.section.anchor ?? '')
const children = computed(() => props.section['array-sections'] ?? [])
const label = computed(() =>
  props.section.line.replace(/<\/?[^>]+>/ig, ' ').trim()
)
const number = computed(() => props.section.number)

const isInView = computed(
  () => !!anchor.value && props.targetIds.includes(anchor.value)
)

const onClick = (e: MouseEvent) => {
  e.preventDefault()
  if (!anchor.value) return
  props.onNavigate(anchor.value)
}
</script>

<template>
  <a :href="`#${anchor}`" :class="{ 'border-[#888]': isInView && props.showRail, 'border-l-2': props.showRail }"
    class="flex w-full items-start gap-1 z-text2 hover:no-underline hover:text-(--link)"
    :style="{ paddingLeft: `calc((${props.depth} + 1) * 0.75rem)` }" @click="onClick">
    <span class="shrink-0 z-text4">
      <span>{{ number }}</span>
      <span v-if="depth === 0">.</span>
    </span>
    <span class="flex-1">{{ label }}</span>
  </a>

  <ul v-if="children?.length > 0" class="p-0 list-none" role="list">
    <li v-for="s in children" :key="s.index ?? s.anchor" class="m-0">
      <TocNode :section="s" :target-ids="targetIds" :depth="props.depth + 1" :show-rail="props.showRail"
        :on-navigate="props.onNavigate" />
    </li>
  </ul>
</template>
