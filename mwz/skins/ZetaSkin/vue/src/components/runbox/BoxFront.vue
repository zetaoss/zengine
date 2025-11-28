<!-- BoxFront.vue -->
<script setup lang="ts">
import { computed, ref } from 'vue'
import type { Job } from './types'

import { SandboxFrame, SandboxConsole, type SandboxLog } from '@common/components/sandbox'

const props = defineProps<{
  job: Job
  seq: number
}>()

const { job, seq } = props

const htmlCode = computed(() => {
  const merged = job.boxes
    .filter((b) => b.lang === 'html')
    .map((b) => b.text)
    .join('\n')
  if (merged.length === 0) return `<html><body></body></html>`
  return merged
})

const jsCode = computed(() => {
  return job.boxes
    .filter((b) => b.lang === 'javascript')
    .map((b) => b.text)
    .join('\n')
})

const logs = ref<SandboxLog[]>([])
const updateLogs = (newLogs: SandboxLog[]) => {
  logs.value = newLogs
}

const sandboxRef = ref<InstanceType<typeof SandboxFrame> | null>(null)
// const run = () => sandboxRef.value?.run()
</script>

<template>
  <div class="p-4 border rounded bg-zinc-100 dark:bg-zinc-900">
    <slot />

    <div v-if="job.main === seq">
      <SandboxFrame ref="sandboxRef" :html="htmlCode" :js="jsCode" class="w-full h-32 border bg-white"
        :class="{ hidden: !job.boxes.some((b) => b.lang === 'html') }" @update:logs="updateLogs" />

      <div v-if="logs.length > 0" class="mt-2">
        <SandboxConsole :logs="logs" />
      </div>
    </div>
  </div>
</template>
