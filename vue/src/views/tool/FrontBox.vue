<!-- FrontBox.vue -->
<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'

import ZButton from '@common/ui/ZButton.vue'
import ConsoleApex from '@common/components/console/ConsoleApex.vue'
import type { Log } from '@common/components/console/types'

declare global {
  interface Window {
    __sandboxLog?: (log: Log) => void
  }
}

const htmlCode = ref('<h1>Hello, World!</h1>')
const jsCode = ref(`console.log('hello');`);
const iframe = ref<HTMLIFrameElement | null>(null)
const logs = ref<Log[]>([])

const handleSandboxLog = (log: Log) => {
  logs.value.push(log)
}

function getContent(): string {
  const raw = htmlCode.value.trim()
  const hasHtmlTag = /^<html[\s>]/i.test(raw)
  let baseHtml: string
  if (hasHtmlTag) {
    baseHtml = raw
  } else {
    baseHtml = `<html><body>${raw}</body></html>`
  }

  const escapedJs = jsCode.value
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
  ['log','info','warn','error','debug'].forEach(function (level) {
    proxy[level] = function () {
      send(level, arguments);
    };
  });
  window.console = proxy;
  try {
    new Function(\`${escapedJs}\`)();
  } catch (e) {
    console.error('Uncaught ' + e.name + ': ' + e.message);
  }
})();
<\/script>`

  if (/<\/body>/i.test(baseHtml)) return baseHtml.replace(/<\/body>/i, scriptBlock + '\n</body>')
  if (/<\/html>/i.test(baseHtml)) return baseHtml.replace(/<\/html>/i, scriptBlock + '\n</html>')
  return baseHtml + scriptBlock
}

const run = () => {
  logs.value = []
  if (!iframe.value) return
  iframe.value.srcdoc = getContent()
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
  <div class="grid grid-cols-1 md:grid-cols-2 gap-4 p-4">
    <div>
      <div class="py-2 flex items-center">
        <b>Playground</b>
        <ZButton class="ml-2" @click="run">Run</ZButton>
      </div>

      <div class="space-y-2">
        <textarea v-model="htmlCode" class="h-[35vh] w-full resize-none font-mono border rounded p-2" />
        <textarea v-model="jsCode" class="h-[35vh] w-full resize-none font-mono border rounded p-2" />
      </div>
    </div>

    <div class="flex flex-col gap-2">
      <div class="h-[50vh] border rounded overflow-hidden">
        <iframe ref="iframe" class="w-full h-full border-0" />
      </div>

      <div class="flex-1 flex flex-col min-h-0">
        <header class="text-center font-bold bg-slate-400 dark:bg-slate-600 text-white py-1">
          Console
        </header>
        <div class="h-[30vh] overflow-y-auto bg-[var(--console-bg)]">
          <ConsoleApex :logs="logs" />
        </div>
      </div>
    </div>
  </div>
</template>
