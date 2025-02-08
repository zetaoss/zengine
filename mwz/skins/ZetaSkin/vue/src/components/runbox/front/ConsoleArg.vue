<script setup lang="ts">
import { onMounted, ref } from 'vue';
import ArgFunction from './ArgFunction.vue'
import ArgObject from './ArgObject.vue';

const props = defineProps<{
  arg: unknown
  depth: number
  expand: boolean
  summary: boolean
}>();

interface Display {
  type: string
  arg: unknown
}

const d = ref<Display>({ type: '', arg: '' })

function parse(arg: unknown): Display {
  if (typeof arg === 'function') {
    return { type: 'function', arg: arg };
  }
  if (typeof arg === 'number') {
    return { type: 'number', arg: arg };
  }
  if (typeof arg === 'object') {
    return { type: 'object', arg: arg }
  }
  if (typeof arg === 'string') {
    if (props.depth == 0) return { type: 'raw', arg: arg }
    else return { type: 'string', arg: `'${arg}'` }
  }
  if (typeof arg === 'symbol') {
    return { type: 'symbol', arg: arg.toString() };
  }
  return { type: 'unknown', arg: `(${typeof arg})${arg}` }
}

onMounted(() => {
  d.value = parse(props.arg)
})
</script>

<template>
  <span :class="d.type">
    <template v-if="d.type == 'function'">
      <ArgFunction :arg="d.arg" />
    </template>
    <template v-else-if="d.type == 'object'">
      <ArgObject :arg="d.arg" :depth="depth" :expand="expand" :summary="summary" />
    </template>
    <template v-else>
      <span :class="d.type">{{ d.arg }}</span>
    </template>
  </span>
</template>

<style scoped lang="scss">
.dict,
.list,
.function {
  @apply italic;
}

.element {
  @apply text-indigo-500 dark:text-indigo-300;
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
