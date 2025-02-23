<script setup lang="ts">
import TheArg from './TheArg.vue';
import { type Log } from './types';

defineProps<{
  logs: Log[]
}>();
</script>

<template>
  <template v-for="(log, i) in logs" :key="i">
    <div :class="log.level">
      <span v-for="(arg, j) in log.args" :key="j">
        <span v-if="j > 0">&nbsp;</span>
        <TheArg :arg="arg" :depth="0" />
      </span>
    </div>
  </template>
</template>

<style scoped>
.system {
  @apply bg-gray-300 text-white dark:bg-gray-700 dark:text-gray-400;
}

.error {
  @apply bg-red-100 dark:bg-[#300] border;
}

.warn {
  @apply bg-yellow-100 dark:bg-[#442];
}
</style>
