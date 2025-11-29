<script setup lang="ts">
import { computed, ref, watch, nextTick } from 'vue'
import type { Job } from './types'
import { SandboxFrame, SandboxConsole, type SandboxLog } from '@common/components/sandbox'

const props = defineProps<{
  job: Job
  seq: number
}>()

const { job, seq } = props

const getCode = (lang: string) => job.boxes
  .filter((b) => b.lang === lang)
  .map((b) => b.text)
  .join('\n')
  .trim()

const htmlCode = computed(() => getCode('html'))
const jsCode = computed(() => getCode('javascript'))

const sandboxId = `sandbox-${job.id}-${seq}`

const logs = ref<SandboxLog[]>([])
const updateLogs = (newLogs: SandboxLog[]) => {
  logs.value = newLogs
}

const sandboxRef = ref<InstanceType<typeof SandboxFrame> | null>(null)
const consoleRef = ref<HTMLElement | null>(null)

watch(
  () => logs.value.length,
  async () => {
    await nextTick()
    const el = consoleRef.value
    if (!el) return
    el.scrollTop = el.scrollHeight
  }
)
</script>

<template>
  <slot />
  <div v-if="job.main === seq">
    <SandboxFrame ref="sandboxRef" v-show="htmlCode.length > 0" :id="sandboxId" :html="htmlCode" :js="jsCode"
      :resizable="job.outResize" class="mt-1 h-32 rounded"
      :class="{ hidden: !job.boxes.some((b) => b.lang === 'html') }" @update:logs="updateLogs" />
    <div v-if="logs.length > 0" class="max-h-40 overflow-y-auto mt-1 bg-[var(--console-bg)]" ref="consoleRef">
      <SandboxConsole :logs="logs" class="rounded-lg" />
    </div>
  </div>
</template>
