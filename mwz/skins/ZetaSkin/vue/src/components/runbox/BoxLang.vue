<script lang='ts'>
import { inject, computed } from 'vue';
import { type Job } from './types';

export default {
  setup() {
    const job = inject<Job>('job');
    const outs = computed(() => job?.resp?.outs ?? []);

    return { outs };
  },
};
</script>

<template>
  <div class="border p-2">
    <slot />
  </div>
  <div v-if="outs.length" class="border font-mono p-2">
    <div v-for="(out, index) in outs" :key="index">
      <div :class="{ 'text-red-500': out.charAt(0) === '2' }">
        {{ out.slice(1) }}
      </div>
    </div>
  </div>
</template>
