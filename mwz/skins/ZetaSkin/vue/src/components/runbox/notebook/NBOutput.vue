<!-- NBOutput.vue -->
<script setup lang="ts">
import { computed } from 'vue';
import { type Output } from '../types';
import './ansi.css';
import './dataframe.css';

const props = defineProps<{
  out: Output
}>()

const ansi2html = (input: string): string => {
  let open = false
  const escaped = input.replace(/&/g, '&amp;').replace(/</g, '&lt;')
  const converted = escaped.replace(/\x1b\[([\d;]*)m/g, (_match, codesStr: string) => {
    const codes = codesStr ? codesStr.split(';').map(c => parseInt(c, 10) || 0) : [0]
    let html = ''
    for (const code of codes) {
      if (code === 0) {
        if (open) {
          html += '</span>'
          open = false
        }
        continue
      }
      if (open) html += '</span>'
      html += `<span class="ansi${code}">`
      open = true
    }
    return html
  })
  return converted + (open ? '</span>' : '')
}

const textPlain = computed(() => props.out.data?.['text/plain']?.join('') ?? '')
const textHtml = computed(() => props.out.data?.['text/html']?.join('') ?? '')
const textLatex = computed(() => props.out.data?.['text/latex']?.join('') ?? '')
const textJson = computed(() => props.out.data?.['application/json']?.join('') ?? '');
</script>

<template>
  <div class="bg-[var(--console-bg)] px-4 py-2 text-sm font-mono rounded-lg"
    :class="{ 'bg-[var(--nbout-error)]': out.output_type == 'error' }">

    <template v-if="out.output_type === 'display_data'">
      <img v-if="out.data?.['image/png']" :src="'data:image/png;base64,' + out.data['image/png']"
        class="bg-white my-2 rounded max-h-[30rem] max-w-[75vw]" />

      <div v-else-if="textHtml" v-html="textHtml"></div>
      <pre v-else-if="textPlain" class="whitespace-pre-wrap">{{ textPlain }}</pre>
      <pre v-else-if="textJson" class="bg-gray-100 dark:bg-gray-900 p-2 rounded overflow-auto">{{ textJson }}</pre>
      <div v-else-if="textLatex" class="text-purple-700 dark:text-purple-300">\({{ textLatex }}\)</div>
    </template>

    <template v-else-if="out.output_type === 'execute_result'">
      <div v-if="textHtml" v-html="textHtml"></div>
      <pre v-else-if="textPlain" class="whitespace-pre-wrap">{{ textPlain }}</pre>
      <div v-else-if="textLatex" class="text-purple-700 dark:text-purple-300">\({{ textLatex }}\)</div>
      <pre v-else-if="textJson" class="bg-gray-100 dark:bg-gray-900 p-2 rounded overflow-auto">{{ textJson }}</pre>
    </template>

    <template v-else-if="out.output_type === 'stream'">
      <pre class="whitespace-pre-wrap">{{ out.text?.join('') }}</pre>
    </template>

    <template v-else-if="out.output_type === 'error'">
      <pre v-html="ansi2html(out.traceback?.join('\n') ?? '')"></pre>
    </template>

    <template v-else>
      <pre class="text-red-500">[Unsupported output_type: {{ out.output_type }}]</pre>
    </template>

  </div>
</template>
