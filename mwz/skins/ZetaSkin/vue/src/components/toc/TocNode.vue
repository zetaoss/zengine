<!-- TocNode.vue -->
<script setup lang="ts">
import type { PropType } from 'vue'
import { computed } from 'vue'
import type { Section } from './types'

defineOptions({ name: 'TocNode' })

const props = defineProps({
  section: { type: Object as PropType<Section>, required: true },
  targetIds: { type: Array as PropType<string[]>, required: true },
  depth: { type: Number, default: 0 },
})

const emit = defineEmits<{ (e: 'navigate', id: string): void }>()

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
  emit('navigate', anchor.value)
  history.pushState(null, '', `#${anchor.value}`)
}
</script>

<template>
  <a :href="`#${anchor}`" :class="{ 'border-[#888] bg-[#8881]': isInView }"
    class="block w-full border-l-2 pl-2 z-muted hover:no-underline hover:text-[var(--link)]" @click="onClick">
    <span class="opacity-50" :style="{ paddingLeft: `calc(${props.depth} * 0.75rem)` }">{{ number }}</span>
    {{ label }}
  </a>

  <ul v-if="children?.length > 0" class="p-0" role="list">
    <li v-for="s in children" :key="s.index ?? s.anchor" class="m-0">
      <TocNode :section="s" :target-ids="targetIds" :depth="props.depth + 1" @navigate="$emit('navigate', $event)" />
    </li>
  </ul>
</template>
