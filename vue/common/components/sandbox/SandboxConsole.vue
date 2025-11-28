<!-- SandboxConsole.vue -->
<script setup lang="ts">
import ConsoleLog from './ConsoleLog.vue'

import type { SandboxLog } from './types'

defineProps<{
  logs: SandboxLog[]
}>()
</script>

<template>
  <div class="console space-y-1 font-mono text-sm">
    <ConsoleLog v-for="(log, i) in logs" :key="i" v-bind="log" />
  </div>
</template>

<style>
.console {
  .string {
    color: var(--console-string);
  }

  .boolean,
  .number {
    color: var(--console-number);
  }

  .null,
  .undefined {
    color: var(--console-null);
  }

  .circular {
    color: var(--console-circular);
    font-style: italic;
  }

  .warn {
    background-color: var(--console-warn-bg);
  }

  .error {
    background-color: var(--console-error-bg);
  }

  .detail {
    padding: 0.25rem;
    margin: 0.25rem 0;
    font-size: 0.875rem;
    line-height: 1.25rem;
    border-radius: 0.25rem;
    background-color: var(--console-detail-bg);
  }

  .arrkey {
    color: var(--console-arrkey);
  }

  .objkey {
    color: var(--console-objkey);
  }
}
</style>
