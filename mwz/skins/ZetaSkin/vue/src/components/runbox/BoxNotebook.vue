<script lang='ts' setup>
import { inject, computed } from 'vue';
import { type Job } from './types';

const job = inject<Job>('job');
const seq = inject<number>('seq') ?? 0;
const outputs = computed(() => job?.outs[seq] ?? []);

const ansi2html = (s: string) => s.replace(/</g, '&lt;').replace(/\x1b\[(\d+;)?(\d+)m/g, (_, _1, code) => code ? `<span class="ansi${code}">` : '</span>') + '</span>';
</script>

<template>
  <div class="border p-2">
    <slot />
  </div>
  <div v-if="outputs.length" class="outputs">
    <div v-for="(o, i) in outputs" :key="i" class="output p-2" :class="o.output_type">
      <template v-if="o.output_type === 'stream'">
        <div v-for="(text, j) in o.text" :key="j">{{ text }}</div>
      </template>
      <template v-else-if="o.output_type === 'display_data' && o.data?.['image/png']">
        <img :src="'data:image/png;base64,' + o.data['image/png']" class="bg-white" />
      </template>
      <template v-else-if="o.output_type === 'execute_result'">
        <div v-if="o.data?.['text/html']" v-html="o.data['text/html'].join('')"></div>
        <pre v-else-if="o.data?.['text/plain']">{{ o.data['text/plain'].join('') }}</pre>
      </template>
      <pre v-else-if="o.output_type === 'error'" v-html="ansi2html(o.traceback?.join('\n') ?? '')" />
    </div>
  </div>
</template>

<style>
.outputs {
  font-family: "Roboto", "Noto", sans-serif;
  @apply text-sm border;
}

.dataframe {
  @apply border;
}

.dataframe th {
  @apply px-2;
}

.dataframe tbody tr {
  text-align: right;
}

.dataframe tbody tr:nth-child(odd) {
  background: #8884;
}

.output.error {
  @apply bg-red-100 dark:bg-red-950/80;
}

.ansi30 {
  @apply text-black dark:text-gray-500;
}

.ansi31 {
  @apply text-red-600 dark:text-red-400;
}

.ansi32 {
  @apply text-green-600 dark:text-green-400;
}

.ansi33 {
  @apply text-yellow-600 dark:text-yellow-400;
}

.ansi34 {
  @apply text-blue-600 dark:text-blue-400;
}

.ansi35 {
  @apply text-purple-600 dark:text-purple-400;
}

.ansi36 {
  @apply text-cyan-600 dark:text-cyan-400;
}

.ansi37 {
  @apply text-gray-600 dark:text-gray-400;
}

.ansi90 {
  @apply text-gray-500 dark:text-gray-600;
}

.ansi91 {
  @apply text-red-400 dark:text-red-300;
}

.ansi92 {
  @apply text-green-400 dark:text-green-300;
}

.ansi93 {
  @apply text-yellow-400 dark:text-yellow-300;
}

.ansi94 {
  @apply text-blue-400 dark:text-blue-300;
}

.ansi95 {
  @apply text-purple-400 dark:text-purple-300;
}

.ansi96 {
  @apply text-cyan-400 dark:text-cyan-300;
}

.ansi97 {
  @apply text-gray-400 dark:text-gray-200;
}
</style>
