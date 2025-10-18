<!-- @/components/CapSticky.vue -->
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useWindowScroll, useWindowSize, onClickOutside, useResizeObserver } from '@vueuse/core'
import { mdiMenu } from '@mdi/js'
import BaseIcon from '@common/ui/BaseIcon.vue'
import { useIsMobile } from '@common/composables/useIsMobile'

const props = defineProps({
  marginY: { type: Number, default: 0 },
  showToggle: { type: Boolean, default: false },
  widthValue: { type: String, default: '100%' },
})

const rootRef = ref<HTMLElement | null>(null)
const collapsed = ref(false)

const { y } = useWindowScroll()
const { height: winH } = useWindowSize()
const { isMobile } = useIsMobile()

const isOverlay = computed(() => props.showToggle && isMobile.value)
const capTop = computed(() => (isMobile.value ? 96 : 48))
const capBottom = computed(() => (isMobile.value ? 0 : 152))

const docH = ref(0)
const readDocH = () => {
  docH.value = document.documentElement.scrollHeight
}

onMounted(() => {
  collapsed.value = isOverlay.value ? true : false
  readDocH()
  useResizeObserver(document.documentElement, readDocH)
})

const styleVars = computed(() => {
  const w = collapsed.value ? 0 : props.widthValue
  const top = Math.max(0, capTop.value - y.value) + props.marginY

  const remain = Math.max(0, docH.value - (y.value + winH.value))
  const bottom = Math.max(0, capBottom.value - remain) + props.marginY

  const h = `calc(100vh - ${top + bottom}px)`
  return {
    width: String(w),
    top: `${top}px`,
    height: h,
    maxHeight: h,
  }
})

const classVars = computed(() => [
  'flex-none shrink-0 z-30 transition-[width]',
  isOverlay.value ? 'fixed' : 'sticky',
])

onClickOutside(rootRef, () => {
  if (isOverlay.value && !collapsed.value) collapsed.value = true
})
</script>

<template>
  <div ref="rootRef" :class="classVars" :style="styleVars">
    <button v-if="showToggle"
      class="absolute p-2 z-10 flex rounded-r-lg opacity-75 bg-gray-200 hover:bg-gray-300 dark:bg-neutral-700 dark:hover:bg-neutral-600 right-0 translate-x-full"
      @click="collapsed = !collapsed">
      <BaseIcon :path="mdiMenu" />
    </button>

    <div class="c-scrollbar h-full overflow-y-auto">
      <slot />
    </div>
  </div>
</template>
