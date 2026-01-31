<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const props = defineProps<{
  modelValue: string
  placeholder?: string
  id: string
  maxHeight?: number
}>()

const textareaEl = ref<HTMLTextAreaElement | null>(null)

function adjustHeight() {
  const el = textareaEl.value
  if (!el) return

  const max = props.maxHeight ?? 200

  el.style.height = 'auto'
  const scrollHeight = el.scrollHeight
  const newHeight = Math.min(scrollHeight, max)

  el.style.height = `${newHeight}px`
  el.style.overflowY = scrollHeight > max ? 'auto' : 'hidden'
}

function handleInput(event: Event) {
  const target = event.target as HTMLTextAreaElement
  emit('update:modelValue', target.value)
  nextTick(adjustHeight)
}

watch(() => props.modelValue, () => { nextTick(adjustHeight) }, { immediate: true })
</script>

<template>
  <textarea ref="textareaEl" class="w-full p-2 bg-inherit resize-none outline-none overflow-y-hidden" :id="id"
    :value="modelValue" :placeholder="placeholder" @input="handleInput" />
</template>
