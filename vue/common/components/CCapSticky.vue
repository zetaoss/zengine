<!-- @common/components/CCapSticky.vue -->
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useWindowScroll, useWindowSize, useEventListener } from '@vueuse/core'

const props = defineProps({
  marginY: { type: Number, default: 0 },
  classes: { type: String, default: '' },
})

const capTop = 48
const capBottom = 152

const docH = ref(0)
const { y } = useWindowScroll()
const { height: winH } = useWindowSize()

const readDocH = () => {
  docH.value = document.documentElement.scrollHeight
}

const styleVars = computed(() => {
  const top = Math.max(0, capTop - y.value) + props.marginY
  const remain = Math.max(0, docH.value - (y.value + winH.value))
  const bottom = Math.max(0, capBottom - remain) + props.marginY

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
  <div class="sticky scrollbar-simple" :class="classes" :style="styleVars">
    <slot />
  </div>
</template>
