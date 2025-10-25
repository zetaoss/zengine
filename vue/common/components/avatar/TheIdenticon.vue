<!-- eslint-disable vue/multi-word-component-names -->
<!-- TheIdenticon.vue -->
<script setup lang="ts">
import { computed } from 'vue'
import { Md5 } from 'ts-md5'

const props = defineProps({
  name: { type: String, required: true },
  size: { type: Number, default: 18 },
})

const data = computed(() => {
  const m = Md5.hashStr(props.name)
  const bytes = m.match(/(.{1,2})/g) || []

  const bits = bytes.map(b => (b[0] > '4' ? 1 : 0))
  const rev = [...bytes].reverse()
  const color = '#' + [0, 1, 2].map(i => `${rev[i]?.[0] ?? '0'}0`).join('')

  const cells: { x: number; y: number }[] = []
  for (let y = 0; y < 5; y++) {
    const rowBits3 = bits.slice(y * 3, y * 3 + 3)
    const rowFull = rowBits3.concat(rowBits3.slice(0, 2).reverse())
    rowFull.forEach((on, x) => {
      if (on) cells.push({ x, y })
    })
  }

  return { color, cells }
})
</script>

<template>
  <svg :style="{ width: `${size}px`, height: `${size}px` }" viewBox="0 0 6 6" xmlns="http://www.w3.org/2000/svg"
    shape-rendering="crispEdges">
    <rect v-for="(c, i) in data.cells" :key="i" :x="c.x + 0.5" :y="c.y + 0.5" width="1" height="1" :fill="data.color" />
  </svg>
</template>
