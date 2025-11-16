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
  <div :class="{ 'is-in-view': isInView }" :style="{ '--depth': props.depth }">
    <div class="border-l-2 pl-2" :class="{ 'border-[#888] bg-[#aaa1]': isInView }">
      <a :href="`#${anchor}`" class="toc-link" @click="onClick">
        <span class="toc-number">{{ number }}</span>
        <span class="toc-label">{{ label }}</span>
      </a>
    </div>

    <ul v-if="children?.length > 0" class="toc-children list-none m-0 p-0" role="list">
      <li v-for="s in children" :key="s.index ?? s.anchor" class="m-0">
        <TocNode :section="s" :target-ids="targetIds" :depth="props.depth + 1" @navigate="$emit('navigate', $event)" />
      </li>
    </ul>
  </div>
</template>

<style scoped>
.toc-node.is-in-view .toc-rail::after {
  content: '';
  position: absolute;
  left: 0;
  top: 0.1rem;
  bottom: 0.1rem;
  width: 2px;
  background-color: var(--link);
  border-radius: 999px;
}

.toc-link {
  flex: 1;
  display: inline-flex;
  align-items: baseline;
  padding: 0.15rem 0;
  text-decoration: none;
  color: var(--muted);
  font-weight: 400;
  padding-left: calc(var(--depth, 0) * 0.75rem);
}

.toc-number {
  opacity: 0.5;
  margin-right: 0.25rem;
  font-variant-numeric: tabular-nums;
}

.toc-label {
  transition: color 0.18s ease-out;
}

.toc-node.is-in-view .toc-label {
  color: var(--text);
}

.toc-children {
  margin-left: 0;
}
</style>
