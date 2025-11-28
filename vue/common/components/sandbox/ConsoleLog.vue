<!-- ConsoleLog.vue -->
<script setup lang="ts">
import ConsoleArg from './ConsoleArg.vue'
import type { LogLevel } from './types'

const props = defineProps<{
  level: LogLevel
  args: unknown[]
}>()

const levelEmoji: Record<LogLevel, string> = {
  log: '',
  error: '❌',
  warn: '⚠️',
  info: 'ℹ️',
  debug: '🐞',
  trace: '🔍',
}

const emoji = levelEmoji[props.level] ?? ''
</script>

<template>
  <div class="my-1 py-1 px-2 rounded border" :class="level">
    <div class="flex gap-2 items-start">
      <span class="w-5 text-center">
        <span v-if="emoji">{{ emoji }}</span>
        <span v-else class="inline-block opacity-0 w-4">&nbsp;</span>
      </span>
      <ConsoleArg v-for="(arg, i) in args" :key="i" :value="arg" />
    </div>
  </div>
</template>
