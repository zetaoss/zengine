<script setup lang="ts">
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
</script>

<template>
  <ul>
    <li v-for="s in toc['array-sections'] || []" :key="s.index">
      <TocSection targetId="" :section="s" />
    </li>
  </ul>
</template>
