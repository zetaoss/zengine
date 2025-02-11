<script setup lang="ts">
import { onMounted, ref } from 'vue';
import TheToggle from './TheToggle.vue';

const props = defineProps<{
  k?: string | number | null;
  arg: unknown
  depth: number
  summary?: boolean
}>();

const typ = ref('')
const text = ref('')
const expandable = ref(false);
const len = ref(0);
const expanded = ref(false);
const obj = ref({});

function inspect(arg: unknown) {
  typ.value = typeof arg
  if (typeof arg === 'object') {
    if (arg === null) {
      typ.value = 'null'
      text.value = 'null'
      return
    }
    obj.value = arg
    len.value = Object.keys(arg).length
    if (Array.isArray(arg)) {
      typ.value = 'array'
    }
    if (!props.summary) {
      expandable.value = true
    }
    return
  }
  if (typeof arg === 'string') {
    typ.value = 'string'
    text.value = arg
    return
  }
  if (typeof arg === 'symbol') {
    typ.value = 'symbol'
    text.value = arg.toString()
    return
  }
  if (typeof arg === 'number') {
    typ.value = 'number'
    text.value = arg.toString()
    return
  }
  text.value = `${arg}`
}

onMounted(() => {
  inspect(props.arg);
});
</script>

<template>
  <span class="inline-block align-top">
    <span class="console-arg" :class="{ 'expand-toggle': expandable }" @click="expanded = !expanded">
      <span v-if="k !== undefined">
        <span>
          <TheToggle :isOpen="expanded" :class="{ invisible: !expandable }" />
        </span>
        <span class="key text-blue-400">{{ k }}:&nbsp;</span>
      </span>
      <span v-else>
        <span v-if="expandable">
          <TheToggle :isOpen="expanded" />
        </span>
      </span>
      <span v-if="depth < 2 || (expandable && !expanded)">({{ len }})&nbsp;</span>
      <span :class="typ">
        <template v-if="typ === 'array'">
          <template v-if="depth > 1 && (summary || expanded)">
            <span>Array({{ len }})</span>
          </template>
          <template v-else>
            <span>[</span>
            <span v-for="(v, k) in obj" :key="k">
              <ConsoleArg :arg="v" :depth="depth + 1" :summary="true" />
              <span v-if="k !== len - 1">, </span>
            </span>
            <span>]</span>
          </template>
        </template>
        <template v-else-if="typ === 'object'">object</template>
        <template v-else>{{ text }}</template>
      </span>
    </span>
    <div v-if="expandable && expanded" class="children">
      <div v-for="(v, k) in obj" :key="k" :style="{ paddingLeft: `${0.5 * depth + 0.5}rem` }">
        <ConsoleArg :k="k" :arg="v" :depth="depth + 1" />
      </div>
    </div>
  </span>
</template>

<style scoped lang="scss">
.expand-toggle {
  cursor: pointer;
  @apply pr-2;
}

.children {
  @apply mt-1;
}

.key {
  @apply text-blue-400;
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
