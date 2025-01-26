<script setup lang="ts">
import { inject } from 'vue'

import BoxFront from './box/front/BoxFront.vue'
import BoxNone from './box/BoxNone.vue'
import BoxNotebook from './box/BoxNotebook.vue'
import BoxRun from './box/BoxRun.vue'
import { type Job, JobType } from './types'

const job = inject('job') as Job
const seq = inject('seq') as number

const getComponent = () => {
  switch (job.type) {
    case JobType.Front: return BoxFront
    case JobType.Notebook: return BoxNotebook
    case JobType.Run: return BoxRun
    default: return BoxNone
  }
}
</script>

<template>
  <component :is="getComponent()" :job="job" :seq="seq">
    <slot />
  </component>
</template>
