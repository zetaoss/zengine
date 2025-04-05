<script setup lang="ts">
import ArgSummary from './ArgSummary.vue';
import { type Log } from './utils';

defineProps<{
  logs: Log[]
}>();
</script>

<template>
  <div class="console">
    <template v-for="(log, i) in logs" :key="i">
      <div class="border" :class="log.level">
        <template v-for="(arg, j) in log.args" :key="j">
          <ArgSummary :level="log.level" :arg="arg" />
        </template>
      </div>
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
}
</style>
