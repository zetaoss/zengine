<script setup lang="ts">
import type { PropType } from 'vue'

import stripTags from '@/utils/str'

import { type Section } from './types'

defineProps({
  section: { type: {} as PropType<Section>, required: true },
  targetId: { type: String, required: true },
})
</script>
<template>
  <div>
    <div class="py-[2px] px-4">
      <a :href="`#${section.anchor}`"
        class="w-full hover:no-underline hover:text-sky-400 inline-block leading-4 z-break-word"
        :class="[section.anchor == targetId ? 'text-z-text' : 'text-z-text2']">
        {{ stripTags(section.line).trim() }}</a>
    </div>
    <ul class="pl-3 py-0 list-none" v-if="section['array-sections'].length > 0">
      <li class="m-0" v-for="s in section['array-sections'] " :key="s.index">
        <TocSection :targetId='targetId' :section="s" />
      </li>
    </ul>
  </div>
</template>
