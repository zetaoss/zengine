<script setup lang="ts">
import TheLog from './TheLog.vue';
import { type Log } from './utils';

defineProps<{
  logs: Log[]
}>();
</script>

<template>
  <div class="console">
    <template v-for="(log, i) in logs" :key="i">
      <TheLog :level="log.level" :args="log.args" />
    </template>
  </div>
</template>

<style>
.console {
  .error {
    @apply bg-red-100 dark:bg-[#300] border;
  }

  .warn {
    @apply bg-yellow-100 dark:bg-[#442];
  }

  .circular {
    @apply text-red-400;
  }

  .number,
  .boolean {
    @apply text-violet-400;
  }

  .null,
  .undefined,
  .graykey {
    @apply text-gray-500;
  }

  .string {
    @apply text-sky-500;
  }

  .bluekey {
    @apply text-blue-400;
  }

  .detail {
    @apply bg-slate-100 dark:bg-slate-700;
  }
}
</style>
