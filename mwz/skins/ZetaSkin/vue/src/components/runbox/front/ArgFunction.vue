<script setup lang="ts">
import { ref, watchEffect } from 'vue'

const props = defineProps<{
  arg: unknown
}>()

const arrow = ref(false)
const text = ref('')

watchEffect(() => {
  if (typeof props.arg === 'function') {
    arrow.value = !('prototype' in props.arg)
    text.value = props.arg.toString().replace(/^function /, '')
  }
})
</script>

<template>
  <span v-if="text">
    <span v-if="!arrow" class="text-orange-400 dark:text-orange-400">ƒ&nbsp;</span>
    <span>{{ text }}</span>
  </span>
</template>
