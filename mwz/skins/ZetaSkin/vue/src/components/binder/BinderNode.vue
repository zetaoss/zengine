<script setup lang="ts">
import type { PropType } from 'vue'
import type { BinderNodeType } from './types'

defineProps({
  node: { type: {} as PropType<BinderNodeType>, required: true },
  depth: { type: Number, default: 0 },
})
</script>
<template>
  <div>
    <div v-if="node.href">
      <a :href="`${node.href}`" :class="{ new: node.new }">{{ node.text }}</a>
    </div>
    <div v-else>
      <span>{{ node.text }}</span>
    </div>
    <ul class="pl-3 py-0 list-none" v-if="node.nodes && node.nodes.length > 0">
      <li class="m-0" v-for="n in node.nodes" :key="n.text">
        <BinderNode :node='n' />
      </li>
    </ul>
  </div>
</template>
