<script setup lang="ts">
import { defineProps } from 'vue'

import type { Word } from './types'

const props = defineProps<{
  word: Word
  noPadding?: boolean
}>()

const { word, noPadding = false } = props
const { words = [], entries = [] } = word
</script>

<template>
  <template v-if="word.typ == 'array'">
    <span class="array">[<template v-for="(w, i) in words" :key="i">
        <ConsoleWord :noPadding="true" :word="w" /><span v-if="i < words.length - 1">, </span>
      </template>]</span>
  </template>
  <template v-else-if="word.typ == 'function'">
    <span class="function"><span class="orange">f</span> {{ word.text }}</span>
  </template>
  <template v-else-if="word.typ == 'object'">
    <span class="object">{<template v-for="(e, i) in entries" :key="i">
        <span class="neutral">{{ e[0] }}</span>:
        <ConsoleWord :noPadding="true" :word="e[1]" /><span v-if="i < entries.length - 1">, </span>
      </template>}</span>
  </template>
  <template v-else>
    <span :class="word.typ">{{ word.text }}</span>
  </template>
  <span v-if="!noPadding">&nbsp;</span>
</template>

<style scoped lang="scss">
.array,
.function,
.object {
  @apply italic;
}

.element {
  @apply text-indigo-500 dark:text-indigo-300;
}

.orange {
  @apply text-orange-400 dark:text-orange-400;
}

.number,
.boolean {
  @apply text-violet-500 dark:text-violet-400;
}

.string,
.symbol {
  @apply text-sky-500 dark:text-sky-300;
}

.neutral,
.undefined {
  @apply text-neutral-400 dark:text-neutral-500;
}
</style>
