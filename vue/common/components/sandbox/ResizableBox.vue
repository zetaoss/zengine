<!-- ResizableBox.vue -->
<script setup lang="ts">
import { ref, useAttrs, computed } from 'vue'
import type { CSSProperties } from 'vue'
import { useEventListener } from '@vueuse/core'

const minPx = 200

const attrs = useAttrs()
const container = ref<HTMLElement | null>(null)

const marginPx = ref(0)

const isDragging = ref(false)
let startX = 0
let startMargin = 0
let maxMargin = 0

const contentStyle = computed<CSSProperties>(() => ({
  marginRight: `${marginPx.value}px`,
  boxSizing: 'border-box',
}))

const onMouseMove = (e: MouseEvent) => {
  if (!isDragging.value || !container.value) return

  const rect = container.value.getBoundingClientRect()
  maxMargin = Math.max(0, rect.width - minPx)

  const delta = startX - e.clientX
  let next = startMargin + delta

  if (next < 0) next = 0
  if (next > maxMargin) next = maxMargin

  marginPx.value = next
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
  maxMargin = Math.max(0, rect.width - minPx)

  startMargin = marginPx.value
}
</script>

<template>
  <div ref="container" v-bind="attrs" class="relative w-full max-w-full">
    <div class="h-full pr-4" :style="contentStyle">
      <slot />
    </div>

    <div v-if="isDragging" class="absolute inset-0 z-10 cursor-col-resize" />

    <div class="absolute inset-y-0 z-20 flex w-4 cursor-col-resize select-none
         items-center justify-center bg-white hover:bg-slate-200 text-sm text-slate-400 rounded-r"
      :style="{ right: `${marginPx}px` }" @mousedown.stop.prevent="onHandleDown">
      ◀
    </div>
  </div>
</template>
