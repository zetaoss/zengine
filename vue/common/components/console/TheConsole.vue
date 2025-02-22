<script setup lang="ts">
import ConsoleArg from './ConsoleArg.vue';
import { type Log } from './types';

defineProps<{
  logs: Log[]
}>();
</script>

<template>
  <template v-for="(log, i) in logs" :key="i">
    <div :class="log.level" v-for="(arg, j) in log.args" :key="j">
      <ConsoleArg :arg="arg" :depth="0" :minify="0" :expandable="true" />
    </div>
  </template>
</template>

<style scoped>
.error {
  @apply bg-red-100 dark:bg-[#533];
}

.warn {
  @apply bg-yellow-100 dark:bg-[#442];
}
</style>
