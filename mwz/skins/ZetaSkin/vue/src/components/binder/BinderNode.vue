<!-- BinderNode.vue -->
<script setup lang="ts">
import { computed, ref } from 'vue'
import type { PropType } from 'vue'
import type { BinderNodeType } from './types'
import BaseIcon from '@common/ui/BaseIcon.vue'
import { mdiTriangleDown } from '@mdi/js'

const props = defineProps({
  node: { type: Object as PropType<BinderNodeType>, required: true },
  depth: { type: Number, default: 0 },
  wgTitle: { type: String, default: '' },
})

const isLink = computed(() => !!props.node.href)
const hasChildren = computed(() => !!props.node.nodes?.length)
const collapsed = ref(props.depth > 0)
const toggle = () => { collapsed.value = !collapsed.value }
</script>

<template>
  <li>
    <div class="flex items-stretch">
      <button v-if="hasChildren" class="w-4 grid place-items-center text-gray-500 rounded hover:bg-gray-500/20"
        @click.stop.prevent="toggle">
        <BaseIcon :path="mdiTriangleDown" class="w-[9px] transition-transform origin-center"
          :class="{ '-rotate-90': collapsed }" />
      </button>
      <div v-else class="w-4">
      </div>
      <component :is="isLink ? 'a' : 'div'" :href="node.href || undefined"
        class="flex-1 rounded hover:bg-gray-500/20 hover:no-underline p-0.5" :class="node.new ? 'new' : ''">
        {{ node.text }}
      </component>
    </div>
    <ul v-if="hasChildren" v-show="!collapsed"
      class="p-0 m-0 pl-2 list-none border-slate-200/60 dark:border-slate-700/40">
      <BinderNode v-for="n in node.nodes" :key="n.text" :node="n" :depth="depth + 1" :wgTitle="wgTitle" />
    </ul>
  </li>
</template>
