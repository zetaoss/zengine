<script setup lang="ts">
import { ref, watch } from 'vue'

const emit = defineEmits(['close'])
const props = defineProps({
  show: { type: Boolean, default: false },
  content: { type: String, default: 'tooltip' },
  direction: { type: String, default: 'top' },
  color: { type: String, default: 'bg-slate-400 dark:bg-slate-700 text-white' },
  timeout: { type: Number, default: 0 },
})

const showTooltip = ref(false)

let timeoutID: ReturnType<typeof setTimeout>

function close() {
  showTooltip.value = false
  emit('close')
}

function updateShow() {
  clearTimeout(timeoutID)
  if (!props.show) {
    close()
    return
  }
  showTooltip.value = true
  if (props.timeout > 0) {
    timeoutID = setTimeout(() => {
      close()
    }, props.timeout)
  }
}

updateShow()
watch(() => props.show, updateShow)
</script>

<template>
  <div class="relative inline-block z-10">
    <slot />
    <div class="absolute" :class="[direction, showTooltip ? '' : 'hidden']">
      <div class="text relative rounded-lg p-1 px-2 text-xs text-center z-20" :class="color">
        {{ content }}
      </div>
      <div class="arrow absolute w-2 h-2 rotate-45" :class="color" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.bottom {
  @apply left-1/2;

  .text {
    @apply -left-1/2 top-[0.5rem];
  }

  .arrow {
    @apply top-0 translate-y-1/2 -translate-x-[0.25rem];
  }
}

.top {
  @apply left-1/2 bottom-full -translate-y-[0.5rem];

  .text {
    @apply -left-1/2;
  }

  .arrow {
    @apply bottom-0 translate-y-1/2 -translate-x-[0.25rem];
  }
}
</style>
