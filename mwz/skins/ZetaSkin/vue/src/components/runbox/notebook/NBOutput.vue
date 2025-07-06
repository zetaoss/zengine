<!-- File: mwz/skins/ZetaSkin/vue/src/components/runbox/Notebook/NBOutput.vue -->
<script setup lang="ts">
import { computed } from 'vue';

import { type Output } from '../types';

const props = defineProps<{ out: Output }>();

const ansi2html = (s: string) => s.replace(/</g, '&lt;').replace(/\x1b\[(\d+;)?(\d+)m/g, (_, _1, code) => code ? `<span class="ansi${code}">` : '</span>') + '</span>';

const textPlain = computed(() => props.out.data?.['text/plain']?.join('') ?? '');
const textHtml = computed(() => props.out.data?.['text/html']?.join('') ?? '');
const textLatex = computed(() => props.out.data?.['text/latex']?.join('') ?? '');
const textJson = computed(() => props.out.data?.['application/json']?.join('') ?? '');
</script>

<template>
  <div class="output p-2 text-sm font-sans" :class="out.output_type">

    <template v-if="out.output_type === 'display_data'">
      <img v-if="out.data?.['image/png']" :src="'data:image/png;base64,' + out.data['image/png']"
        class="bg-white max-w-full my-2 rounded" />

      <div v-else-if="textHtml" v-html="textHtml"></div>

      <pre v-else-if="textPlain" class="whitespace-pre-wrap">{{ textPlain }}</pre>

      <pre v-else-if="textJson" class="bg-gray-100 dark:bg-gray-900 p-2 rounded overflow-auto">
        {{ textJson }}
      </pre>

      <div v-else-if="textLatex">
        <span class="font-mono text-purple-700 dark:text-purple-300">\({{ textLatex }}\)</span>
      </div>
    </template>

    <template v-else-if="out.output_type === 'execute_result'">
      <div v-if="textHtml" v-html="textHtml"></div>
      <pre v-else-if="textPlain" class="whitespace-pre-wrap">{{ textPlain }}</pre>
      <div v-else-if="textLatex">
        <span class="font-mono text-purple-700 dark:text-purple-300">\({{ textLatex }}\)</span>
      </div>
      <pre v-else-if="textJson" class="bg-gray-100 dark:bg-gray-900 p-2 rounded overflow-auto">
        {{ textJson }}
      </pre>
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

<style lang="scss">
@use './ansi.scss';
@use './dataframe.scss';
</style>

<style lang="scss" scoped>
.output.error {
  @apply bg-red-100 dark:bg-red-950/80;
}

img {
  @apply max-h-[30rem] max-w-[75vw];
}
</style>
