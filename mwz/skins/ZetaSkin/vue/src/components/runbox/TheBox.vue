<script setup lang="ts">
import { inject } from 'vue'

import BoxFront from './front/BoxFront.vue'
import BoxLang from './BoxLang.vue'
import BoxNotebook from './BoxNotebook.vue'
import BoxZero from './BoxZero.vue'
import { type Job, JobType } from './types'

const job = inject('job') as Job
const seq = inject('seq') as number

const getComponent = () => {
  switch (job.type) {
    case JobType.Front: return BoxFront
    case JobType.Lang: return BoxLang
    case JobType.Notebook: return BoxNotebook
    default: return BoxZero
  }
}
</script>

<template>
  <component :is="getComponent()" :job="job" :seq="seq">
    <slot />
  </component>
</template>
