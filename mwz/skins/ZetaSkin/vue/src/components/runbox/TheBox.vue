<script setup lang="ts">
import { inject } from 'vue'

import BoxFront from './BoxFront.vue'
import BoxLang from './BoxLang.vue'
import BoxNotebook from './BoxNotebook.vue'
import BoxZero from './BoxZero.vue'

import type { Job } from './types'

const job = inject('job') as Job
const seq = inject('seq') as number
</script>

<template>
  <div v-if="job">
    <div v-if="job.type == 'front'">
      <BoxFront :job="job" :seq="seq">
        <slot />
      </BoxFront>
    </div>
    <div v-else-if="job.type == 'lang'">
      <BoxLang :job="job" :seq="seq">
        <slot />
      </BoxLang>
    </div>
    <div v-else-if="job.type == 'notebook'">
      <BoxNotebook :job="job" :seq="seq">
        <slot />
      </BoxNotebook>
    </div>
    <div v-else>
      <BoxZero job="job" :seq="0">
        <slot />
      </BoxZero>
    </div>
  </div>
</template>
