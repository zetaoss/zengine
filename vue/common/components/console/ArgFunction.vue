<script setup lang="ts">
import { ref, watchEffect } from 'vue'

const props = defineProps<{
  arg: unknown
  inEntry: boolean
}>()

const arrow = ref(false)
const text0 = ref('')
const text1 = ref('')

watchEffect(() => {
  if (typeof props.arg === 'function') {
    const s = props.arg.toString()
    arrow.value = !('prototype' in props.arg) && !s.startsWith("function")
    text0.value = s.replace(/^function /, '')
    text1.value = s.replace(/^function /, '').replace(/ { \[native code\] }$/, '').replace(/{CONSOLE}$/, '')
  }
})
</script>

<template>
  <template v-if="!inEntry">
    <span>ƒ</span>
  </template>
  <template v-else>
    <span v-if="!arrow" class="text-orange-400 dark:text-orange-400">ƒ&nbsp;</span>
    <span>{{ text1 }}</span>
  </template>
  <!-- <template v-else-if="minify == 1">
    <span v-if="!arrow" class="text-orange-400 dark:text-orange-400">ƒ&nbsp;</span>
    <span>{{ text1 }}</span>
  </template>  -->
</template>
