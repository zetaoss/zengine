<!-- eslint-disable vue/multi-word-component-names -->
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
    <template v-if="job.type == 'front'">
      BoxFront
      <BoxFront :job="job" :seq="seq">
        <slot />
      </BoxFront>
    </template>
    <template v-else-if="job.type == 'lang'">
      <BoxLang :job="job" :seq="seq">
        <slot />
      </BoxLang>
    </template>
    <template v-else-if="job.type == 'notebook'">
      BoxNotebook
      <BoxNotebook :job="job" :seq="seq">
        <slot />
      </BoxNotebook>
    </template>
    <template v-else>
      <BoxZero :job="job" :seq="seq">
        <slot />
      </BoxZero>
    </template>
  </div>
</template>
