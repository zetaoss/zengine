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

const getCode = (lang: string) => job.boxes.filter((b) => b.lang === lang).map((b) => b.text).join('\n').trim()
const htmlCode = computed(() => getCode('html'))
const jsCode = computed(() => getCode('javascript'))

const sandboxId = `sandbox-${job.id}-${seq}`
console.log(sandboxId)

const logs = ref<SandboxLog[]>([])
const updateLogs = (newLogs: SandboxLog[]) => {
  logs.value = newLogs
}

const sandboxRef = ref<InstanceType<typeof SandboxFrame> | null>(null)
</script>

<template>
  <div class="p-4 border rounded bg-zinc-100 dark:bg-zinc-900">
    <slot />

    <div v-if="job.main === seq">
      <SandboxFrame ref="sandboxRef" v-show="htmlCode.length > 0" :id="sandboxId" :html="htmlCode" :js="jsCode"
        :resizable="job.outResize" class="w-full h-32 border bg-white"
        :class="{ hidden: !job.boxes.some((b) => b.lang === 'html') }" @update:logs="updateLogs" />

      <div v-if="logs.length > 0" class="mt-2">
        <SandboxConsole :logs="logs" />
      </div>
    </div>
  </div>
</template>
