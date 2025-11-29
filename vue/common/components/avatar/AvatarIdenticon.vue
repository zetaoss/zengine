<!-- AvatarIdenticon.vue -->
<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps({
  name: { type: String, required: true },
  size: { type: Number, default: 18 },
})

const hash = (s: string) => s.split('').reduce((h, c) => (h ^ c.charCodeAt(0)) * -5, 5) >>> 2

const identicon = computed(() => {
  const h = hash(props.name)
  const hue = (h % 9) * (360 / 9)
  const color = `hsl(${hue} 95% 45%)`

  const cells = Array.from({ length: 25 }, (_, i) => {
    if (!(h & (1 << (i % 15)))) return null
    const col = (i / 5) | 0
    const row = i % 5
    const x = col < 3 ? col : 7 - col
    const y = row
    return { x, y }
  }).filter(Boolean) as { x: number; y: number }[]

  return { color, cells }
})
</script>

<template>
  <svg :width="size" :height="size" viewBox="0 0 12 12" xmlns="http://www.w3.org/2000/svg" shape-rendering="crispEdges"
    role="img" :aria-label="`Identicon for ${props.name}`" class="bg-white">
    <rect v-for="(c, i) in identicon.cells" :key="i" :x="c.x * 2 + 1" :y="c.y * 2 + 1" width="2" height="2"
      :fill="identicon.color" />
  </svg>
</template>
