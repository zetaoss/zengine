<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'

import { mdiClose } from '@mdi/js'

import ZIcon from '@common/ui/ZIcon.vue'
import ZButton from '@common/ui/ZButton.vue'

const props = defineProps({
  show: { type: Boolean, required: true },
  title: { type: String, default: '' },
  okText: { type: String, default: '확인' },
  okColor: { type: String, default: 'danger' },
  okDisabled: { type: Boolean, default: false },
  cancelText: { type: String, default: '취소' },
})

const emit = defineEmits(['ok', 'cancel'])

function keyup(event: KeyboardEvent) {
  if (event.key === 'Escape') {
    emit('cancel')
  }
}

function clickOK() {
  if (props.okDisabled) return
  emit('ok')
}

onMounted(() => {
  document.addEventListener('keyup', keyup)
})
onUnmounted(() => {
  document.removeEventListener('keyup', keyup)
})
</script>

<template>
  <div class="fixed table z-40 top-0 left-0 w-full h-full transition bg-[#000a]" :class="show ? '' : 'hidden'">
    <div class="table-cell align-middle">
      <div class="w-full max-w-[60vw] md:max-w-[40vw] border rounded bg-white dark:bg-gray-900 transition m-auto">
        <div class="relative">
          <div class="absolute right-0">
            <button type="button" class="w-8 h-8" @click="emit('cancel')">
              <ZIcon :path="mdiClose" />
            </button>
          </div>
        </div>
        <div class="flex w-full items-start justify-between p-5 border-b rounded-t">
          <slot />
        </div>
        <hr class="border-0">
        <div class="p-3 flex justify-center gap-3">
          <ZButton :disabled="okDisabled" :color="okColor" @click="clickOK()">
            {{ okText }}
          </ZButton>
          <ZButton @click="emit('cancel')">
            {{ cancelText }}
          </ZButton>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.btn {
  @apply rounded mx-0.5 text-xs p-2 px-4;
  background-color: var(--z-btn);
  color: var(--z-text);
  transition: 80ms cubic-bezier(0.33, 1, 0.68, 1);
  transition-property: color, background-color, box-shadow, border-color;

  &:hover {
    @apply no-underline brightness-90 dark:brightness-150 text-inherit;
  }

  &:disabled,
  &.disabled {
    @apply opacity-50 cursor-not-allowed;
  }
}
</style>
