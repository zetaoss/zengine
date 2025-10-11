<!-- BinderNode.vue -->
<script setup lang="ts">
import { type PropType, computed, ref, onMounted, nextTick, watchEffect } from 'vue'
import type { BinderNodeType } from './types'
import BaseIcon from '@common/ui/BaseIcon.vue'
import { mdiTriangleDown } from '@mdi/js'
import { useStorage } from '@vueuse/core'

const props = defineProps({
  node: { type: Object as PropType<BinderNodeType>, required: true },
  depth: { type: Number, default: 0 },
  wgTitle: { type: String, default: '' },
  binderId: { type: Number, required: true },
  parentPath: { type: String, default: '' },
  idx: { type: Number, default: 0 },
})
const emit = defineEmits<{ (e: 'reveal'): void }>()

const isCurrent = computed(() => props.node.text === props.wgTitle)
const isLink = computed(() => !!props.node.href)
const hasChildren = computed(() => !!props.node.nodes?.length)

const key = `${props.parentPath}/${props.idx}`
const storageKey = computed(() => `binder-${props.binderId}`)
const expandedMap = useStorage<Record<string, number>>(storageKey, {}, localStorage)
const expanded = ref(props.depth === 0)
const rowRef = ref<HTMLElement | null>(null)

watchEffect(() => {
  const saved = expandedMap.value[key]
  if (saved === 0 || saved === 1) expanded.value = !!saved
})

function persist() {
  expandedMap.value[key] = expanded.value ? 1 : 0
}

function toggle() {
  expanded.value = !expanded.value
  persist()
}

function handleReveal() {
  if (!expanded.value) {
    expanded.value = true
    persist()
  }
  emit('reveal')
}

async function centerScrollIfCurrent() {
  if (!isCurrent.value) return
  emit('reveal')
  await nextTick()
  rowRef.value?.scrollIntoView({ block: 'center', behavior: 'smooth' })
}

onMounted(centerScrollIfCurrent)
</script>

<template>
  <li>
    <div class="flex items-stretch" ref="rowRef">
      <button v-if="hasChildren" class="w-4 grid place-items-center text-gray-500 rounded hover:bg-gray-500/20"
        @click.stop.prevent="toggle">
        <BaseIcon :path="mdiTriangleDown" class="w-[9px] transition-transform origin-center"
          :class="{ '-rotate-90': !expanded }" />
      </button>
      <div v-else class="w-4"></div>

      <component :is="isLink ? 'a' : 'div'" :href="node.href || undefined"
        class="flex-1 rounded hover:bg-gray-500/20 hover:no-underline p-0.5"
        :class="isCurrent ? 'font-bold' : node.new ? 'new' : ''">
        {{ node.text }}
      </component>
    </div>

    <ul v-if="hasChildren" v-show="expanded"
      class="p-0 m-0 pl-2 list-none border-slate-200/60 dark:border-slate-700/40">
      <BinderNode v-for="(n, i) in node.nodes" :key="n.text" :node="n" :depth="depth + 1" :wgTitle="wgTitle"
        :binderId="binderId" :parentPath="key" :idx="i" @reveal="handleReveal" />
    </ul>
  </li>
</template>
