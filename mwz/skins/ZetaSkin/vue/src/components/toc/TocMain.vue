<script setup lang="ts">
import { ref } from 'vue'
import type { PropType } from 'vue'

import type { Section } from './types'
import TocSection from './TocSection.vue'

const props = defineProps({
  toc: {
    type: [Object, String] as PropType<Section | string>,
    required: true,
  },
})

const toc = props.toc as Section

const count = ref(0)
const increment = () => count.value++
</script>

<template>
  <ul>
    <li>
      <button @click="increment">Click me</button>
      <span>Count: {{ count }}</span>
    </li>
    <li>toc: {{ toc }}</li>
    <li v-for="s in toc['array-sections'] || []" :key="s.index">
      <TocSection targetId="" :section="s" />
    </li>
  </ul>
</template>
