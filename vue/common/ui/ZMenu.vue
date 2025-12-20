<!-- @common/ui/ZMenu.vue -->
<script setup lang="ts">
import { useDismissable } from '@common/composables/useDismissable'
import { computed, getCurrentInstance, ref } from 'vue'

const props = defineProps<{
  modelValue?: boolean
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const instance = getCurrentInstance()

const isControlled = computed(() => {
  const vp = instance?.vnode.props ?? {}
  return 'onUpdate:modelValue' in vp
})

const rootEl = ref<HTMLElement | null>(null)
const innerOpen = ref(false)

const open = computed({
  get: () => (isControlled.value ? !!props.modelValue : innerOpen.value),
  set: (v: boolean) => {
    if (isControlled.value) emit('update:modelValue', v)
    else innerOpen.value = v
  },
})

function close() {
  open.value = false
}

function toggle() {
  if (props.disabled) return
  open.value = !open.value
}

function show() {
  if (props.disabled) return
  open.value = true
}

useDismissable(rootEl, {
  enabled: open,
  onDismiss: close,
})

defineExpose({ show, close, toggle })
</script>

<template>
  <div ref="rootEl" class="relative inline-flex">
    <slot name="trigger" :open="open" :toggle="toggle" :close="close" />

    <div v-if="open"
      class="absolute top-full right-0 z-50 mt-1 py-1 inline-flex w-max flex-col rounded bg-[var(--z-card-bg)] ring-1 ring-[var(--z-border)] shadow-lg"
      role="menu" @click.stop>
      <slot name="menu" :close="close" />
    </div>
  </div>
</template>
