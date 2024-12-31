<script setup lang="ts">
import { onMounted, ref } from 'vue'

const props = defineProps({
  name: { type: String, required: true },
  size: { type: Number, default: 18 },
})
const letter = ref('')
const background = ref('')
const dy = ref('30%')
const ratio = ref(0.8)

onMounted(() => {
  const n = `${props?.name}`.toUpperCase().split(' ')
  letter.value = (n.length === 1) ? (n[0] ? n[0].charAt(0) : '?') : (n[0].charAt(0) + n[1].charAt(0))
  if (letter.value.length > 1) {
    ratio.value = 0.5
    dy.value = '20%'
  }
  const s = `${props?.name}abcdef`
  let a = 0
  for (let i = 0; i < s.length; i++) a = (a << 5) - a + s.charCodeAt(i)
  const b = ((a.toString(16).match(/[a-f0-9]{6}$/g) || [])[0]?.match(/.{1,2}/g) || []).map((e) => parseInt(e, 16))
  background.value = `hsla(${(360 * b[0]) / 256},${(60 * b[1]) / 256 + 40}%,${(50 * b[2]) / 256 + 30}%, 0.8)`
})
</script>
<template>
  <svg class="w-full h-full" :style="`background:${background}`">
    <text text-anchor="middle" x="50%" y="50%" :dy="dy" fill="#fff" :font-size="size * ratio" font-family="Arial">
      {{ letter }}
    </text>
  </svg>
</template>
