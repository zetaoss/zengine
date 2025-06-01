<!-- eslint-disable vue/multi-word-component-names -->
<script setup lang="ts">
import LogVue from './Log.vue'
import type { Log } from './utils'

defineProps<{
  logs: Log[]
}>()
</script>

<template>
  <div class="console space-y-1 font-mono text-sm">
    <LogVue v-for="(log, i) in logs" :key="i" :level="log.level" :args="log.args" :initialExpanded="true" />
  </div>
</template>

<style>
.console {
  .string {
    @apply text-sky-500;
  }

  .number,
  .boolean {
    @apply text-violet-400;
  }

  .null,
  .undefined {
    @apply text-gray-500;
  }

  .circular {
    @apply text-red-400 italic;
  }

  .log {
    @apply text-white;
  }

  .warn {
    @apply bg-yellow-100 dark:bg-yellow-900;
  }

  .error {
    @apply bg-red-100 dark:bg-red-900;
  }

  .detail {
    @apply bg-slate-100 dark:bg-slate-800 rounded p-1 my-1 text-sm;
  }

  .graykey {
    @apply text-gray-400;
  }

  .bluekey {
    @apply text-blue-400;
  }
}
</style>
