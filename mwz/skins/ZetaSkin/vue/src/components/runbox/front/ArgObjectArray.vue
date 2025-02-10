<script setup lang="ts">
import { ref, watchEffect } from 'vue'
import ConsoleArg from './ConsoleArg.vue'
import TheItem from './TheItem.vue'

const props = defineProps<{
  arg: unknown
  depth: number
  expand: boolean
  expandable: boolean
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
    <template v-if="!expandable">
      Array({{ arr.length }})
    </template>
    <template v-else>
      <span>[</span>
      <template v-for="(x, i) in arr" :key="i">
        <ConsoleArg :arg="x" :depth="depth + 1" :expandable="false" />
        <span v-if="i < arr.length - 1">, </span>
      </template>
      <span>]</span>
    </template>
    <template v-if="isExpanded">
      <template v-for="(x, i) in arr" :key="i">
        <TheItem :arg="x" :idx="i" :depth="depth + 1" :expandable="false" />
      </template>
    </template>
  </span>
</template>
