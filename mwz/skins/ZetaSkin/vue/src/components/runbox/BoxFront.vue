<!-- BoxFront.vue -->
<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue'
import type { Job } from './types'

import ConsoleApex from '@common/components/console/ConsoleApex.vue'
import type { Log } from '@common/components/console/types'

declare global {
  interface Window {
    __sandboxLog?: (log: Log) => void
  }
}

const props = defineProps<{
  job: Job
  seq: number
}>()

const { job, seq } = props

const iframe = ref<HTMLIFrameElement | null>(null)
const logs = ref<Log[]>([])

const handleSandboxLog = (log: Log) => {
  logs.value.push(log)
}

function getContentFromJob(job: Job): string {
  const htmlParts = job.boxes
    .filter((b) => b.lang === 'html')
    .map((b) => b.text.trim())
    .filter(Boolean)

  let baseHtml = htmlParts.join('\n')

  const hasHtmlTag = /^<html[\s>]/i.test(baseHtml)

  if (!baseHtml) baseHtml = '<html><body></body></html>'
  else if (!hasHtmlTag) baseHtml = `<html><body>${baseHtml}</body></html>`

  const jsCode = job.boxes
    .filter((b) => b.lang === 'javascript')
    .map((b) => b.text)
    .join('\n')

  const escapedJs = jsCode
    .replace(/\\/g, '\\\\')
    .replace(/`/g, '\\`')
    .replace(/\$\{/g, '\\${')

  const scriptBlock = `
<script>
(function () {
  function send(level, args) {
    var argArray = Array.prototype.slice.call(args);
    if (window.parent && typeof window.parent.__sandboxLog === 'function') {
      window.parent.__sandboxLog({ level: level, args: argArray });
    }
  }

  var proxy = {};
  ['log', 'info', 'warn', 'error', 'debug'].forEach(function (level) {
    proxy[level] = function () {
      send(level, arguments);
    };
  });

  window.console = proxy;

  try {
    new Function(\`${escapedJs}\`)();
  } catch (e) {
    console.error('Uncaught ' + e.name + ': ' + e.message + (e.stack ? ' at ' + e.stack : ''));
  }
})();
<\/script>`

  if (/<\/body>/i.test(baseHtml)) return baseHtml.replace(/<\/body>/i, scriptBlock + '\n</body>')
  if (/<\/html>/i.test(baseHtml)) return baseHtml.replace(/<\/html>/i, scriptBlock + '\n</html>')
  return baseHtml + scriptBlock
}

function run() {
  logs.value = []
  if (!iframe.value) return
  iframe.value.srcdoc = getContentFromJob(job)
}

onMounted(() => {
  window.__sandboxLog = handleSandboxLog
  run()
})

onBeforeUnmount(() => {
  delete window.__sandboxLog
})
</script>

<template>
  <div class="p-4 border rounded bg-zinc-100 dark:bg-zinc-900">
    <slot />
    <div v-if="job.main === seq">
      <iframe ref="iframe" class="w-full h-32 border bg-white"
        :class="{ hidden: !job.boxes.some((b) => b.lang === 'html') }" />

      <div v-if="logs.length > 0" class="mt-2 flex flex-col min-h-0">
        <header class="text-center font-bold bg-slate-400 dark:bg-slate-600 text-white py-1">
          Console
        </header>
        <div class="max-h-40 overflow-y-auto bg-[var(--console-bg)]">
          <ConsoleApex :logs="logs" />
        </div>
      </div>
    </div>
  </div>
</template>
