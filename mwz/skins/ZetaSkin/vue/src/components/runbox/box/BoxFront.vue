<script setup lang="ts">
import { onMounted, ref } from 'vue'

import TheIcon from '@common/components/TheIcon.vue'
import { mdiAlert, mdiCloseCircle } from '@mdi/js'

import type { Job } from '../types'

import ConsoleWord from './front/ConsoleWord.vue'
import type { Line, Word } from './front/types'

declare global {
  interface Window {
    console: any
  }
}

const props = defineProps<{
  job: Job,
  seq: number
}>()
const { job, seq } = props
const iframe = ref<HTMLIFrameElement | null>(null)
const lines = ref<Line[]>([])

function formatArg(arg: any): Word {
  const typ: string = typeof arg
  switch (typ) {
    case 'function':
      return { typ, text: arg.toString().replace(/^function /, '') }
    case 'object':
      if (arg === null) return { typ: 'undefined', text: 'null' }
      if (arg.tagName) return { typ: 'element', text: arg.outerHTML }
      if (Array.isArray(arg)) return { typ: 'array', words: arg.map(formatArg) }
      return { typ, entries: Object.entries(arg).map(([k, v]) => [k, formatArg(v)]) }
    case 'string':
      return { typ, text: `'${arg}'` }
    case 'symbol':
      return { typ, text: arg.toString() }
    case 'undefined':
      return { typ, text: 'undefined' }
    default: // number, boolean
      return { typ, text: `${arg}` }
  }
}

function myConsole(sev: string, args: any[]) {
  const words: Word[] = []
  if (args.length >= 1 && typeof args[0] === 'string') {
    words.push({ typ: 'plain', words: [], text: args.shift() })
  }
  words.push(...args.map((a) => formatArg(a)))
  lines.value.push({ sev, words })
}
onMounted(() => {
  if (!iframe.value) return
  const { contentDocument: doc, contentWindow: win } = iframe.value
  if (!doc || !win) return
  win.console = new Proxy(console, { get(_, prop) { return (...args: any[]) => myConsole(prop as string, args) } })
  doc.open()
  doc.write(job.boxes.map((b) => (b.lang === 'javascript' ? `<script>${b.text}<${'/'}script>` : b.text)).join(''))
  doc.close()
})
</script>

<template>
  <div class="p-4 border rounded bg-zinc-50 dark:bg-zinc-950">
    <slot />
  </div>
  <div v-if='seq === job.main'>
    <iframe ref="iframe" class="w-full border bg-white" :class="{ hidden: !job.boxes.some(b => b.lang === 'html') }"
      title="" />
    <div v-if="lines.length > 0" class="border font-mono text-sm">
      <div v-for="(l, i) in lines" :key="i" class="p-1 pt-0">
        <div class="grid grid-cols-[30px_1fr] rounded" :class="l.sev">
          <div class="text-center">
            <template v-if="l.sev == 'warn'">
              <TheIcon :path="mdiAlert" :size="13" />
            </template>
            <template v-else-if="l.sev == 'error'">
              <TheIcon :path="mdiCloseCircle" :size="13" />
            </template>
          </div>
          <div class="col">
            <template v-for="(w, j) in l.words" :key="j">
              <ConsoleWord :word="w" />
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.log {
  .col {
    @apply border-t;
  }
}

.warn {
  @apply bg-yellow-400 bg-opacity-15 text-orange-400;
}

.error {
  @apply bg-red-400 bg-opacity-15 text-red-400;
}
</style>
