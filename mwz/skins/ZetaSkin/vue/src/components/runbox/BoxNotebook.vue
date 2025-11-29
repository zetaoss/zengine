<!-- BoxNotebook.vue -->
<script setup lang="ts">
import { computed } from 'vue'
import NBOutput from './notebook/NBOutput.vue'
import type { Job } from './types'

const props = defineProps<{
  job: Job
  seq: number
}>()

const { job, seq } = props

const nbouts = computed(() => job.notebookOuts[seq] ?? [])
const loaded = computed(() => job.phase != null)
</script>

<template>
  <slot />

  <div v-if="loaded && job.phase === 'succeeded' && nbouts.length">
    <NBOutput v-for="(nbout, i) in nbouts" :key="i" :out="nbout" />
  </div>
</template>
