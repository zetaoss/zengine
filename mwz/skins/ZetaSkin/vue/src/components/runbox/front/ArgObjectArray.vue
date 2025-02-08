<script setup lang="ts">
import { ref, watchEffect } from 'vue'
import ConsoleArg from './ConsoleArg.vue'

const props = defineProps<{
  arg: unknown
  depth: number
  expand: boolean
  summary: boolean
}>();

const arr = ref<unknown[]>([])
const isExpanded = ref(props.expand)

watchEffect(() => {
  const a = props.arg
  if (Array.isArray(a)) {
    arr.value = a
  }
})
</script>

<template>
  <span>
    <span v-if="!summary" @click="isExpanded = !isExpanded">
      <span v-text="isExpanded ? '▼' : '▶'"></span>
      <span>&nbsp;({{ arr.length }})&nbsp;</span>
    </span>
    <template v-if="summary">
      Array({{ arr.length }})
    </template>
    <template v-else>
      <span>[</span>
      <template v-for="(x, i) in arr" :key="i">
        <ConsoleArg :arg="x" :depth="depth + 1" :expand="false" :summary="true" />
        <span v-if="i < arr.length - 1">, </span>
      </template>
      <span>]</span>
    </template>
    <template v-if="isExpanded">
      <template v-for="(x, i) in arr" :key="i">
        <div>{{ i }}</div>
      </template>
    </template>
  </span>
</template>
