<script setup lang="ts">
import { computed, inject } from 'vue'

import BoxFront from './BoxFront.vue'
import BoxLang from './BoxLang.vue'
import BoxNotebook from './BoxNotebook.vue'
import BoxZero from './BoxZero.vue'

import type { Job } from './types'

const job = inject('job') as Job
const seq = inject('seq') as number

const componentMap = {
  'front': BoxFront,
  'lang': BoxLang,
  'notebook': BoxNotebook,
  '': BoxZero,
} as const

const dynamicComponent = computed(() => {
  if (!job) return BoxZero
  return componentMap[job.type as keyof typeof componentMap] ?? BoxZero
})
</script>

<template>
  <component :is="dynamicComponent" :job="job" :seq="seq">
    <slot />
  </component>
</template>
