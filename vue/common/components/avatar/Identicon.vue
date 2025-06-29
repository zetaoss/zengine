<!-- eslint-disable vue/multi-word-component-names -->
<script setup lang="ts">
import { onMounted, ref, nextTick } from 'vue'
import { Md5 } from 'ts-md5'

const props = defineProps({
  name: { type: String, required: true },
  size: { type: Number, required: true },
})

const el = ref<HTMLCanvasElement | null>(null)

onMounted(async () => {
  await nextTick()

  if (!el.value) return

  const ctx = el.value.getContext('2d')
  if (!ctx) return

  const m = Md5.hashStr(`${props?.name}`)
  const p = m.match(/(.{1,2})/g)
  const q = p!.map((_, i) => (p![i][0] > '4' ? 1 : 0))
  p!.reverse()
  const r = [0, 0, 0]
  const f = `#${r.map((_, i) => `${p![i][0]}0`).join('')}`
  const g = '#0000'
  const a = [0, 0, 0, 0, 0]
  const b = a.map((__, i) => a.map((_, j) => q[i * 3 + j]))

  ctx.scale(60, 30)
  b.forEach((c, i) => {
    c.forEach((_, j) => {
      ctx.fillStyle = j >= 3 ? (c[j === 3 ? 1 : 0] ? f : g) : (c[j] ? f : g)
      ctx.fillRect(j, i, 1, 1)
    })
  })
})
</script>

<template>
  <canvas ref="el" class="w-full h-full p-[5%] bg-zinc-100" />
</template>
