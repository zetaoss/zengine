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
  <div class="p-4 border rounded bg-zinc-100 dark:bg-zinc-900">
    <slot />

    <div v-if="job.main === seq">
      <div class="bg-gray-500/20 rounded">
        <SandboxFrame ref="sandboxRef" v-show="htmlCode.length > 0" :id="sandboxId" :html="htmlCode" :js="jsCode"
          :resizable="job.outResize" class="h-32 rounded" :class="{ hidden: !job.boxes.some((b) => b.lang === 'html') }"
          @update:logs="updateLogs" />
      </div>
      <div v-if="logs.length > 0" class="max-h-40 overflow-y-auto" ref="consoleRef">
        <SandboxConsole :logs="logs" />
      </div>
    </div>
  </div>
</template>
