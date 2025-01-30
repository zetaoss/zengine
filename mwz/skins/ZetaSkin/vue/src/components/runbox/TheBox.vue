<script setup lang="ts">
import { inject } from 'vue'

import BoxFront from './BoxFront.vue'
import BoxNotebook from './BoxNotebook.vue'
import BoxRun from './BoxRun.vue'
import BoxZero from './BoxZero.vue'
import { type Job, JobType } from './types'

const job = inject('job') as Job
const seq = inject('seq') as number

const getComponent = () => {
  switch (job.type) {
    case JobType.Front: return BoxFront
    case JobType.Notebook: return BoxNotebook
    case JobType.Run: return BoxRun
    default: return BoxZero
  }
}
</script>

<template>
  <component :is="getComponent()" :job="job" :seq="seq">
    <slot />
  </component>
</template>
