<!-- @/components/CapSticky.vue -->
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { onClickOutside } from '@vueuse/core'
import { mdiMenu } from '@mdi/js'
import BaseIcon from '@common/ui/BaseIcon.vue'
import { useIsMobile } from '@common/composables/useIsMobile'

const props = defineProps({
  marginY: { type: Number, default: 0 },
  showToggle: { type: Boolean, default: false },
  widthValue: { type: String, default: '100%' },
})

const root = ref<HTMLElement | null>(null)
const collapsed = ref(false)

const { isMobile } = useIsMobile()
const isOverlay = computed(() => props.showToggle && isMobile.value)

onMounted(() => {
  collapsed.value = isOverlay.value ? true : false
})

onClickOutside(root, () => {
  if (isOverlay.value && !collapsed.value) collapsed.value = true
})

const toggle = () => { collapsed.value = !collapsed.value }

const styleVars = computed(() => {
  const top = `calc(var(--navbar-visible-height) + ${props.marginY}px)`
  const marginTop = `${props.marginY}px`
  const height = isOverlay.value
    ? `calc(100vh - (var(--navbar-visible-height) + ${props.marginY * 2}px))`
    : `calc(100vh - (var(--navbar-visible-height) + var(--footer-visible-height) + ${props.marginY * 2}px))`

  return {
    width: collapsed.value ? '0' : props.widthValue,
    marginTop,
    top,
    height,
  }
})
</script>

<template>
  <div ref="root" class="flex-none shrink-0 z-30 transition-[width]" :class="isOverlay ? 'fixed' : 'sticky'"
    :style="styleVars">
    <button v-if="showToggle"
      class="absolute p-2 z-10 flex rounded-r-lg opacity-75 bg-gray-200 hover:bg-gray-300 dark:bg-neutral-700 dark:hover:bg-neutral-600 right-0 translate-x-full"
      :aria-expanded="!collapsed" @click="toggle">
      <BaseIcon :path="mdiMenu" />
    </button>

    <div class="z-scrollbar h-full w-full overflow-y-auto">
      <slot />
    </div>
  </div>
</template>
