<!-- BinderNode.vue -->
<script setup lang="ts">
import type { PropType } from 'vue'
import type { BinderNodeType } from './types'

defineProps({
  node: { type: Object as PropType<BinderNodeType>, required: true },
  depth: { type: Number, default: 0 },
  wgTitle: { type: String, default: '' },
})

const indent = 14
const padX = 8
</script>

<template>
  <li>
    <component :is="node.href ? 'a' : 'div'" :href="node.href || undefined"
      class="block w-full rounded-md px-2 py-1 transition-colors hover:bg-slate-200 dark:hover:bg-slate-700 hover:no-underline"
      :class="[
        node.text === wgTitle ? 'current bg-slate-300/50 dark:bg-slate-700/50' : '',
        node.href ? 'cursor-pointer' : '',
        node.new ? 'text-red-700' : 'text-sky-700 dark:text-sky-500',
      ]" :style="{ paddingLeft: `${padX + depth * indent}px`, paddingRight: `${padX}px` }">
      {{ node.text }}
    </component>

    <ul v-if="node.nodes?.length" class="p-0 m-0 pl-2 list-none border-slate-200/60 dark:border-slate-700/40">
      <BinderNode v-for="n in node.nodes" :key="n.text" :node="n" :depth="depth + 1" :wgTitle="wgTitle" />
    </ul>
  </li>
</template>
