<!-- SandboxFrame.vue -->
<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import type { SandboxLog } from './types'
import buildHtml from './buildHtml'

declare global {
  interface Window {
    __sandboxLog?: (log: SandboxLog) => void
  }
}

const props = defineProps<{
  html: string
  js: string
}>()

const emit = defineEmits<{
  (e: 'update:logs', logs: SandboxLog[]): void
}>()

const iframe = ref<HTMLIFrameElement | null>(null)
const logs = ref<SandboxLog[]>([])

const handleSandboxLog = (log: SandboxLog) => {
  logs.value.push(log)
  emit('update:logs', logs.value)
}

const run = () => {
  logs.value = []
  emit('update:logs', logs.value)

  if (!iframe.value) return
  const content = buildHtml(props.html, props.js)
  iframe.value.srcdoc = content
}

defineExpose({ run })

onMounted(() => {
  window.__sandboxLog = handleSandboxLog
  run()
})

onBeforeUnmount(() => {
  delete window.__sandboxLog
})
</script>

<template>
  <iframe ref="iframe" class="w-full h-full border-0" />
</template>
