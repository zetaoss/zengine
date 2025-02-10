<script setup lang="ts">
import ConsoleArg from './ConsoleArg.vue'

const props = defineProps<{
  idx: number
  arg: unknown
  depth: number
}>();

const a = props.arg
const expandable = typeof a === 'object' && a !== null
</script>

<template>
  <div :style="`padding-left:${depth * 2}rem`">
    <span v-if="expandable">
      ▶
    </span>
    <span class="text-blue-400">{{ idx }}</span>
    <span>:&nbsp;</span>
    <ConsoleArg :arg="arg" :depth="depth" :expandable="expandable" />
  </div>
</template>
