<!-- BinderNode.vue -->
<script setup lang="ts">
import { type PropType, computed, ref, onMounted, nextTick, watchEffect } from 'vue'
import type { BinderNodeType } from './types'
import ZIcon from '@common/ui/ZIcon.vue'
import { mdiChevronRight } from '@mdi/js'
import { useStorage } from '@vueuse/core'

const props = defineProps({
  node: { type: Object as PropType<BinderNodeType>, required: true },
  depth: { type: Number, default: 0 },
  wgArticleId: { type: Number, required: true },
  binderId: { type: Number, required: true },
  parentPath: { type: String, default: '' },
  idx: { type: Number, default: 0 },
})
const emit = defineEmits<{ (e: 'reveal'): void }>()

const isCurrent = computed(() => props.node.id === props.wgArticleId)
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
  rowRef.value?.scrollIntoView({ block: 'center' })
}

onMounted(centerScrollIfCurrent)
</script>

<template>
  <li class="flex flex-col">
    <div class="row flex items-stretch hover:bg-gray-500/15" ref="rowRef">
      <component :is="isLink ? 'a' : 'div'" :href="node.href || undefined" class="flex flex-1 px-2 hover:no-underline"
        :class="{ 'font-bold': isCurrent, 'new': node.new }" :title="node.text">
        <span :style="{ paddingLeft: `${depth}rem` }">{{ node.text }}</span>
      </component>

      <button v-if="hasChildren" class="w-6 h-6 grid place-items-center rounded-full hover:bg-gray-500/30"
        @click.stop.prevent="toggle">
        <ZIcon :path="mdiChevronRight" class="transition-transform" :class="{ 'rotate-90': expanded }" />
      </button>
    </div>

    <ul v-if="hasChildren" v-show="expanded" class="p-0 m-0">
      <BinderNode v-for="(n, i) in node.nodes" :key="n.text" :node="n" :depth="depth + 1" :wgArticleId="wgArticleId"
        :binderId="binderId" :parentPath="key" :idx="i" @reveal="handleReveal" />
    </ul>
  </li>
</template>
