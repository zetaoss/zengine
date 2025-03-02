<script setup lang="ts">
import BriefAny from './BriefAny.vue';
import { type Log } from './utils';

defineProps<{
  logs: Log[]
}>();
</script>

<template>
  <div class="console">
    <template v-for="(log, i) in logs" :key="i">
      <div class="border flex" :class="log.level">
        <div>
          <template v-for="(param, j) in log.params" :key="j">
            <BriefAny :item="param.item" :expandableIfCollection="true" />
          </template>
        </div>
      </div>
    </template>
  </div>
</template>

<style>
.console {
  .system {
    @apply bg-gray-300 text-white dark:bg-gray-700 dark:text-gray-400;
  }

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
