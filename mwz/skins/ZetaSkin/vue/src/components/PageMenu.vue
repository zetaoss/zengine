<!-- PageMenu.vue -->
<script setup lang="ts">
import { useDismissable } from '@common/composables/useDismissable'
import ZIcon from '@common/ui/ZIcon.vue'
import { mdiDotsVertical } from '@mdi/js'
import { ref } from 'vue'

const root = ref<HTMLElement | null>(null)
const show = ref(false)

function close() {
  show.value = false
}

function toggle() {
  show.value = !show.value
}

useDismissable(root, {
  enabled: show,
  onDismiss: close,
})
</script>

<template>
  <div ref="root" class="page-menu relative print:hidden z-link">
    <button type="button" class="page-btn" @click="toggle">
      <ZIcon :path="mdiDotsVertical" />
    </button>

    <div v-if="show"
      class="absolute z-30 right-0 p-0 border list-none rounded bg-white shadow-md dark:bg-neutral-800 text-sm text-black dark:text-white"
      @click.stop>
      <slot />
    </div>
  </div>
</template>
