<script setup lang="ts">
import type { PropType } from 'vue'
import TocSection from './TocSection.vue'

import stripTags from '@/utils/str'

import { type Section } from './types'

const props = defineProps({
  section: { type: {} as PropType<Section>, required: true },
  targetId: { type: String, required: true },
})

const anchor = props.section.anchor ?? ''
const tid = props.targetId ?? ''
</script>
<template>
  <div>
    <div class="py-[2px] px-4">
      <a :href="`#${anchor}`" class="w-full hover:no-underline hover:text-sky-400 inline-block leading-4 z-break-word"
        :class="[anchor == tid ? 'text-z-text' : 'text-z-text2']">
        {{ stripTags(section.line).trim() }}</a>
    </div>
    <ul class="pl-3 py-0 list-none" v-if="section['array-sections'].length > 0">
      <li class="m-0" v-for="s in section['array-sections']" :key="s.index">
        <TocSection :targetId='tid' :section="s" />
      </li>
    </ul>
  </div>
</template>
