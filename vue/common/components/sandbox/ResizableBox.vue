<script setup lang="ts">
import { ref, useAttrs, computed } from 'vue'
import { useEventListener } from '@vueuse/core'

const minPx = 200

const attrs = useAttrs()
const container = ref<HTMLElement | null>(null)
const widthPx = ref<number | null>(null)

const isDragging = ref(false)
let startX = 0
let startWidth = 0

const widthStyle = computed(() => ({
  width: widthPx.value !== null ? `${widthPx.value}px` : '100%',
}))

const onMouseMove = (e: MouseEvent) => {
  if (!isDragging.value) return

  const delta = e.clientX - startX
  let next = startWidth + delta

  if (next > startWidth) next = startWidth
  if (next < minPx) next = minPx

  widthPx.value = next
}

useEventListener(window, 'mousemove', onMouseMove)
useEventListener(window, 'mouseup', () => {
  if (!isDragging.value) return
  isDragging.value = false
})

const onHandleDown = (e: MouseEvent) => {
  if (!container.value) return

  isDragging.value = true
  startX = e.clientX

  const rect = container.value.getBoundingClientRect()
  startWidth = widthPx.value ?? rect.width
}
</script>

<template>
  <div ref="container" v-bind="attrs" class="group relative max-w-full" :style="widthStyle">
    <div class="h-full pr-4">
      <slot />
    </div>
    <div v-if="isDragging" class="absolute inset-0 z-10 cursor-col-resize" />
    <div class="absolute inset-y-0 right-0 z-20 flex w-4 cursor-col-resize select-none items-center justify-center"
      @mousedown.stop.prevent="onHandleDown">
      <div class="h-[40%] w-2 rounded-full bg-slate-950/20 group-hover:bg-slate-950/40" />
    </div>
  </div>
</template>
