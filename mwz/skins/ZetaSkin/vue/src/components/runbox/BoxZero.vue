<script setup lang="ts">
import { computed } from 'vue'
import { type Job } from './types'
import CopyButton from '@common/ui/CopyButton.vue'

const props = defineProps<{
  job: Job
  seq?: number
}>()

const box = computed(() => {
  const seq = props.seq ?? 0
  return props.job?.boxes?.[seq] ?? { lang: '', text: '' }
})
</script>

<template>
  <div class="float-right relative z-10 bg-red-500 p-2 flex items-center gap-x-2 rounded">
    <span class="text-white text-sm font-mono">{{ box.lang }}</span>
    <CopyButton :text="box.text" />
  </div>
  <slot />
</template>
