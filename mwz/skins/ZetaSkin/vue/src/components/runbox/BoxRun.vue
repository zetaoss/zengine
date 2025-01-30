<script lang='ts'>
import { inject, computed } from 'vue';
import { type Job } from './types';

export default {
  setup() {
    const job = inject<Job>('job');
    const logs = computed(() => job?.resp?.logs ?? []);

    return { logs };
  },
};
</script>

<template>
  <div class="border p-2">
    <slot />
  </div>
  <div v-if="logs.length" class="border font-mono p-2">
    <div v-for="(log, index) in logs" :key="index">
      <div :class="{ 'text-red-500': log.charAt(0) === '2' }">
        {{ log.slice(1) }}
      </div>
    </div>
  </div>
</template>
