<script setup lang="ts">
import type { PropType } from 'vue'

import type Item from './types'

defineProps({
  items: { type: Array as PropType<Item[]>, default: [] as Item[], require: true },
})
</script>

<template>
  <div v-for="item in items" :key="item.text">
    <RouterLink
      class="inline-block w-full p-2 px-8 hover:no-underline text-xs text-z-text hover:bg-zinc-300 dark:hover:bg-zinc-700"
      :to="item.href">
      {{ item.text }}
    </RouterLink>
  </div>
</template>
