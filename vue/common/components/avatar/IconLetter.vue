<!-- IconLetter.vue -->
<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps({
  name: { type: String, required: true },
  size: { type: Number, default: 18 },
})

const safeName = computed(() => props.name.trim() || '?')

const initial = computed(() => {
  const parts = safeName.value.toUpperCase().split(/\s+/).filter(Boolean)
  const letters = parts.map(p => p[0] ?? '').join('').slice(0, 2)
  return letters || '?'
})

const bgColor = computed(() => {
  let hash = 5381
  const str = safeName.value

  for (let i = 0; i < str.length; i++) {
    hash = ((hash << 5) + hash) + str.charCodeAt(i)
  }

  const hue = ((hash % 360) + 360) % 360
  return `oklch(60% 0.18 ${hue})`
})

const fontSize = computed(() =>
  initial.value.length > 1 ? 42 : 64
)
</script>

<template>
  <svg :width="size" :height="size" viewBox="0 0 100 100" :style="{ background: bgColor }">
    <text x="50" y="50" fill="#fff" font-weight="800" :font-size="fontSize" text-anchor="middle"
      dominant-baseline="central" dy="-.05em">
      {{ initial }}
    </text>
  </svg>
</template>
