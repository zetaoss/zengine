<!-- CapSticky.vue -->
<script setup lang="ts">
import { useDismissable } from '@common/composables/useDismissable'
import { useIsMobile } from '@common/composables/useIsMobile'
import ZIcon from '@common/ui/ZIcon.vue'
import { mdiMenu } from '@mdi/js'
import { computed, onMounted,ref } from 'vue'

const props = defineProps({
  marginY: { type: Number, default: 0 },
  showToggle: { type: Boolean, default: false },
  widthValue: { type: String, default: '100%' },
  navBlurColor: { type: String, default: '' },
})

const root = ref<HTMLElement | null>(null)
const collapsed = ref(false)

const { isMobile } = useIsMobile()
const isOverlay = computed(() => props.showToggle && isMobile.value)

onMounted(() => {
  collapsed.value = isOverlay.value ? true : false
})

function toggle() {
  collapsed.value = !collapsed.value
}

function close() {
  collapsed.value = true
}

useDismissable(root, {
  enabled: isOverlay,
  onDismiss: close,
})

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
      class="absolute p-2 z-20 flex rounded-r-lg opacity-75 bg-gray-200 hover:bg-gray-300 dark:bg-neutral-700 dark:hover:bg-neutral-600 right-0 translate-x-full"
      :aria-expanded="!collapsed" @click="toggle">
      <ZIcon :path="mdiMenu" />
    </button>

    <div class="z-scrollbar h-full w-full overflow-y-auto">
      <slot />
    </div>
    <footer v-if="navBlurColor" :style="{ backgroundColor: navBlurColor }" />
  </div>
</template>

<style scoped>
footer {
  width: calc(100% - 6px);
  mask-image: linear-gradient(transparent, #000 64px);
  bottom: 0;
  height: 64px;
  left: 0;
  pointer-events: none;
  position: absolute;
}
</style>
