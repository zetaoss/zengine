<script setup lang="ts">
import { onMounted, ref } from 'vue';
import TheToggle from './TheToggle.vue';
import ArgString from './ArgString.vue';
import ArgFunction from './ArgFunction.vue';
import ListEntries from './ListEntries.vue';

const props = withDefaults(
  defineProps<{
    arg: unknown;
    depth: number;
    seen?: Set<unknown>;
    inEntry?: boolean;
    inConsole?: boolean;
  }>(),
  {
    seen: () => new Set(),
    inEntry: false,
    inConsole: false,
  }
);

const expanded = ref(false)
const typ = ref('')
const text = ref('')
const entries = ref<[string | number, unknown][]>([]);
const name = ref('');
const localSeen = ref(new Set(props.seen));

function inspect(arg: unknown) {
  if (localSeen.value.has(arg)) {
    typ.value = 'circular'
    text.value = 'circular'
    return
  }
  typ.value = typeof arg
  if (typeof arg === 'function') {
    typ.value = 'function'
    return
  }
  if (typeof arg === 'object') {
    if (arg === null) {
      typ.value = 'null'
      text.value = 'null'
      return
    }
    localSeen.value.add(arg)
    if (Array.isArray(arg)) {
      typ.value = 'Array'
      entries.value = arg.map((value, index) => [index, value])
      return
    }
    name.value = arg.constructor.name.replace('Object', '')
    if (arg instanceof Map) {
      typ.value = 'Map'
      entries.value = [...arg.entries()]
      return
    }
    if (arg instanceof Set) {
      typ.value = 'Set'
      entries.value = [...arg].map((value, index) => [index, value])
      return
    }
    if (Object.keys(arg).length > 0) {
      typ.value = 'Object'
      entries.value = Object.entries(arg)
      if (name.value === 'console') {
        entries.value = entries.value.map(([name]) => [name, Function(`return function ${name}(){CONSOLE}`)()]);
      }
      return
    }
  }
  if (typeof arg === 'string') {
    text.value = arg
    if (props.depth == 0) {
      typ.value = 'raw'
      return
    }
    return
  }
  if (typeof arg === 'symbol') {
    text.value = arg.toString()
    return
  }
  // number, object, string
  text.value = `${arg}`
}

onMounted(() => {
  inspect(props.arg);
});
</script>

<template>
  <template v-if="entries.length == 0">
    <template v-if="typ == 'function'">
      <ArgFunction :arg="arg" :inEntry="inEntry" :inConsole="inConsole" />
    </template>
    <template v-else-if="typ == 'string'">
      <ArgString :text="text" :inEntry="inEntry" />
    </template>
    <template v-else>
      <span :class="typ">{{ text }}</span>
    </template>
  </template>
  <template v-else>
    <span v-if="depth == 0">
      <TheToggle :isOpen="expanded" @click="expanded = !expanded" />
    </span>
    <template v-if="typ == 'Array'">
      <span>
        <span>[</span>
        <span v-for="(entry, i) in entries?.slice(0, 5)" :key="i">
          <span v-if="i > 0">, </span>
          <TheArg :arg="entry[1]" :depth="depth + 1" :seen="localSeen" :inEntry="inEntry" />
        </span>
        <span v-if="entries?.length > 5">, …</span>
        <span>]</span>
      </span>
    </template>
    <template v-else-if="typ == 'Object'">
      <span class="italic">
        <span v-if="name">{{ name }}</span>
        <span>{</span>
        <span v-for="(entry, i) in entries?.slice(0, 5)" :key="i">
          <span v-if="i > 0">, </span>
          <span class="text-gray-400">{{ entry[0] }}</span>
          <span>:&nbsp;</span>
          <TheArg :arg="entry[1]" :depth="depth + 1" :seen="localSeen" :inEntry="inEntry" />
        </span>
        <span v-if="entries?.length > 5">, …</span>
        <span>}</span>
      </span>
    </template>
    <template v-if="expanded">
      <ListEntries :typ="typ" :depth="depth" :seen="localSeen" :entries="entries" />
    </template>
  </template>
</template>

<style scoped lang="scss">
.children {
  @apply mt-1;
}

.function {
  @apply italic;
}

.number,
.boolean {
  @apply text-violet-500 dark:text-violet-400;
}

.string,
.symbol {
  @apply text-sky-500 dark:text-sky-300;
}

.null,
.undefined {
  @apply text-neutral-400 dark:text-neutral-500;
}
</style>
