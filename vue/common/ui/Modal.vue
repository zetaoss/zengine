<!-- eslint-disable vue/multi-word-component-names -->
<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'

import { mdiClose } from '@mdi/js'

import Icon from '@common/ui/Icon.vue'

const props = defineProps({
  show: { type: Boolean, required: true },
  title: { type: String, default: '' },
  okText: { type: String, default: 'OK' },
  okClass: { type: String, default: '' },
  okDisabled: { type: Boolean, default: false },
  cancelText: { type: String, default: 'Cancel' },
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
              <Icon :path="mdiClose" />
            </button>
          </div>
        </div>
        <div class="flex w-full items-start justify-between p-5 border-b rounded-t">
          <slot />
        </div>
        <hr class="border-0">
        <div class="p-3 text-center rounded-b">
          <button type="button" class="btn" :class="[okDisabled ? 'disabled' : '', okClass]" @click="clickOK()">
            {{ okText }}
          </button>
          <button type="button" class="btn" @click="emit('cancel')">
            {{ cancelText }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
