<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { Job } from './types'
import TheConsole from '@common/components/console/TheConsole.vue'
import type { Log } from '@common/components/console/utils'

declare global {
  interface Window {
    console: Console;
  }
}

const props = defineProps<{
  job: Job
  seq: number
}>()

const { job, seq } = props
const iframe = ref<HTMLIFrameElement | null>(null)
const logs = ref<Log[]>([])

function generateJobScript(job: Job): string {
  return job.boxes.map(b => {
    if (b.lang !== 'javascript') return b.text
    return `<script>
      try {
        new Function(\`${b.text}\`)();
      } catch(e) {
        console.error("Uncaught " + e.name + ": " + e.message + " at " + e.stack);
      }
    <\/script>`
  }).join('')
}

onMounted(() => {
  const doc = iframe.value?.contentDocument
  const win = iframe.value?.contentWindow
  if (!doc || !win) return

  // proxy 콘솔
  win.console = new Proxy(console, {
    get(_, level) {
      return (...args: unknown[]) => {
        logs.value.push({ level: level as string, args })
      }
    }
  })

  // HTML/JS 출력
  doc.open()
  doc.write(generateJobScript(job))
  doc.close()
})
</script>

<template>
  <div class="p-4 border rounded bg-zinc-100 dark:bg-zinc-900">
    <slot />
    <div v-if="job.main === seq">
      <iframe ref="iframe" class="w-full h-32 border bg-white"
        :class="{ hidden: !job.boxes.some(b => b.lang === 'html') }" />
      <div v-if="logs.length > 0" class="border font-mono text-sm p-2 pb-5">
        <TheConsole :logs="logs" />
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.line {
  @apply m-0.5 grid grid-cols-[30px_1fr] rounded border-b;
}

.warn {
  @apply bg-yellow-400 bg-opacity-15 text-orange-400;
}

.error {
  @apply bg-red-400 bg-opacity-15 text-red-400;
}

.log+.log .col {
  @apply border-t pt-0.5;
}
</style>
