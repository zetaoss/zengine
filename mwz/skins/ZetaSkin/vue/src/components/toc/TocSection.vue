<!-- TocSection.vue -->
<script setup lang="ts">
import type { PropType } from 'vue'
import { computed } from 'vue'
import type { Section } from './types'

defineOptions({ name: 'TocSection' })

const props = defineProps({
  section: { type: Object as PropType<Section>, required: true },
  targetId: { type: String, required: true },
})
const emit = defineEmits<{ (e: 'navigate', id: string): void }>()

const anchor = computed(() => props.section.anchor ?? '')
const children = computed(() => props.section['array-sections'] ?? [])
const label = computed(() => props.section.line.replace(/<\/?[^>]+>/ig, ' ').trim())
const number = computed(() => props.section.number)

const isPrimary = computed(() => anchor.value === props.targetId)

const linkClass = computed(() => ['hover:no-underline', 'hover:text-sky-400', isPrimary.value ? 'text-sky-500 semi-bold' : 'text-z-text'])

const onClick = (e: MouseEvent) => {
  e.preventDefault()
  emit('navigate', anchor.value)
  history.pushState(null, '', `#${anchor.value}`)
}
</script>

<template>
  <div>
    <a :href="`#${anchor}`" :aria-current="isPrimary ? 'location' : undefined" :class="linkClass" @click="onClick">
      <span class="opacity-50">{{ number }}</span>
      {{ label }}
    </a>

    <ul v-if="children?.length > 0" class="pl-3 py-0 list-none m-0" role="list">
      <li v-for="s in children" :key="s.index ?? s.anchor" class="m-0">
        <TocSection :targetId="targetId" :section="s" @navigate="$emit('navigate', $event)" />
      </li>
    </ul>
  </div>
</template>
