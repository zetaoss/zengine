<script setup lang="ts">
import { computed } from 'vue'
import ConsoleArg from './ConsoleArg.vue'
import type { LogLevel } from './types'

const props = defineProps<{
  level: LogLevel
  args: unknown[]
}>()

const emoji = computed(() => getComputedStyle(document.documentElement).getPropertyValue(`--console-emoji-${props.level}`))
</script>

<template>
  <div class="px-2 py-0.5 rounded border-b" :class="level">
    <div class="flex gap-2 items-start">
      <span class="w-4 text-center">
        <span v-if="emoji">{{ emoji }}</span>
        <span v-else class="inline-block opacity-0 w-4">&nbsp;</span>
      </span>

      <ConsoleArg v-for="(arg, i) in args" :key="i" :value="arg" />
    </div>
  </div>
</template>
