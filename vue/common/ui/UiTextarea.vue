<script setup lang="ts">
import { ref, watch, onMounted, nextTick } from 'vue'

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const props = defineProps<{
  modelValue: string
  placeholder?: string
  id: string
  maxHeight?: number
}>()

const val = ref(props.modelValue ?? '')
const textareaEl = ref<HTMLTextAreaElement | null>(null)

function adjustHeight(el: HTMLTextAreaElement) {
  const max = props.maxHeight ?? 200
  el.style.height = 'auto'

  const scrollHeight = el.scrollHeight
  const newHeight = Math.min(scrollHeight, max)

  el.style.height = newHeight + 'px'
  if (scrollHeight > max) {
    el.style.overflowY = 'auto'
  } else {
    el.style.overflowY = 'hidden'
  }
}

function handleInput(event: Event) {
  const target = event.target as HTMLTextAreaElement
  emit('update:modelValue', val.value)
  adjustHeight(target)
}

watch(
  () => props.modelValue,
  v => {
    val.value = v ?? ''
    nextTick(() => {
      if (textareaEl.value) adjustHeight(textareaEl.value)
    })
  }
)

onMounted(() => {
  nextTick(() => {
    if (textareaEl.value) adjustHeight(textareaEl.value)
  })
})
</script>

<template>
  <textarea ref="textareaEl" class="bg-transparent active:ring-0 focus:ring-0" v-model="val" :id="id"
    :placeholder="placeholder" required @input="handleInput" />
</template>

<style scoped>
textarea {
  width: 100%;
  resize: none;
  outline: none;
  overflow-y: hidden;
}
</style>
