<!-- SandboxFrame.vue -->
<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, withDefaults } from 'vue'
import ResizableBox from './ResizableBox.vue'
import type { SandboxLog } from './types'
import buildHtml from './buildHtml'

declare global {
  interface Window {
    [key: string]: unknown
  }
}

const props = withDefaults(
  defineProps<{
    id: string
    html: string
    js: string
    resizable?: boolean
  }>(),
  {
    resizable: false,
  },
)

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
  const content = buildHtml(props.id, props.html, props.js)
  iframe.value.srcdoc = content
}

defineExpose({ run })

onMounted(() => {
  window[props.id] = handleSandboxLog
  run()
})

onBeforeUnmount(() => {
  if (window[props.id] === handleSandboxLog) {
    delete window[props.id]
  }
})
</script>

<template>
  <component :is="props.resizable ? ResizableBox : 'div'" class="h-full rounded">
    <iframe ref="iframe" class="w-full h-full border-0 bg-white" />
  </component>
</template>
