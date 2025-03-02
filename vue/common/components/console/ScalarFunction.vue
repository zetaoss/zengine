<script setup lang="ts">
import { ref, watchEffect } from 'vue'
import { type Data } from './utils'

const props = defineProps<{
  data: Data
  mode: number // 0: normal, 1: short, 2: minimal
}>()

const arrow = ref(false)
const text0 = ref('')
const text1 = ref('')

watchEffect(() => {
  const a = props.data.arg;
  if (typeof a === 'function') {
    const s = a.toString()
    arrow.value = !('prototype' in a) && !s.startsWith("function")
    text0.value = s.replace(/^function /, '')
    text1.value = s.replace(/^function /, '').replace(/ { \[native code\] }$/, '').replace(/{CONSOLE}$/, '')
  }
})
</script>

<template>
  <span class="italic">
    <template v-if="mode == 2">
      <span>ƒ</span>
    </template>
    <template v-else>
      <span v-if="!arrow" class="text-orange-400 dark:text-orange-400">ƒ&nbsp;</span>
      <span>{{ text1 }}</span>
    </template>
  </span>
</template>
