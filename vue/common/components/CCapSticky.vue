<!-- @common/components/CCapSticky.vue -->
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useWindowScroll, useWindowSize, useEventListener } from '@vueuse/core'

const capTop = 48
const capBottom = 152
const marginY = 16

const docH = ref(0)
const { y } = useWindowScroll()
const { height: winH } = useWindowSize()

const readDocH = () => {
  docH.value = document.documentElement.scrollHeight
}

const styleVars = computed(() => {
  const top = Math.max(0, capTop - y.value) + marginY
  const remain = Math.max(0, docH.value - (y.value + winH.value))
  const bottom = Math.max(0, capBottom - remain) + marginY

  return {
    top: `${top}px`,
    maxHeight: `calc(100vh - ${top + bottom}px)`,
  }
})

onMounted(() => {
  readDocH()
  useEventListener(window, 'resize', readDocH)
})
</script>

<template>
  <div class="sticky overflow-y-auto scrollbar-simple mr-4" :style="styleVars">
    <slot />
  </div>
</template>
