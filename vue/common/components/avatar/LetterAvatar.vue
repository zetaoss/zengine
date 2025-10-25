<!-- LetterAvatar.vue -->
<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps({
  name: { type: String, required: true },
  size: { type: Number, default: 18 },
})

const initial = computed(() => {
  const trimmed = (props.name || '').trim().toUpperCase().split(' ')
  if (trimmed.length === 0) return '?'
  if (trimmed.length === 1) return trimmed[0]?.charAt(0) || '?'
  return trimmed[0].charAt(0) + trimmed[1].charAt(0)
})

const bgColor = computed(() => generateColor(props.name))

const fontSize = computed(() => (initial.value.length > 1 ? 60 : 80))

function generateColor(name: string): string {
  const seed = `${name}abcdef`
  let hash = 0
  for (let i = 0; i < seed.length; i++) {
    hash = (hash << 5) - hash + seed.charCodeAt(i)
  }

  const matches = hash.toString(16).match(/[a-f0-9]{6}$/g)
  const parts = matches?.[0]?.match(/.{1,2}/g)?.map(hex => parseInt(hex, 16))

  if (!parts || parts.length < 3) {
    return 'hsla(0, 0%, 70%, 0.8)'
  }

  const [r, g, b] = parts
  const h = (360 * r) / 256
  const s = (60 * g) / 256 + 40
  const l = (50 * b) / 256 + 30

  return `hsla(${h}, ${s}%, ${l}%, 0.8)`
}
</script>

<template>
  <svg :width="size" :height="size" viewBox="0 0 100 100" class="font-sans" role="img" aria-hidden="true">
    <rect x="0" y="0" width="100" height="100" rx="20" ry="20" :fill="bgColor" />
    <text x="50" y="45" fill="white" font-weight="600" :font-size="fontSize" text-anchor="middle"
      dominant-baseline="central">
      {{ initial }}
    </text>
  </svg>
</template>
