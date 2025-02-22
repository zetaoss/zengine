<script setup lang="ts">
import { onMounted, ref } from 'vue';
import ArgEntries from './ArgEntries.vue';
import ArgFunction from './ArgFunction.vue';
import ListEntries from './ListEntries.vue';
import TheToggle from './TheToggle.vue';

const props = defineProps<{
  id?: string | number | null;
  arg: unknown
  depth: number
  minify: number
  expandable: boolean
}>();

const typ = ref('')
const text = ref('')
const expanded = ref(false);
const entries = ref<[string | number, unknown][] | null>(null);
const contructor = ref('');

function inspect(arg: unknown) {
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
    if (Array.isArray(arg)) {
      typ.value = 'Array'
      entries.value = arg.map((value, index) => [index, value])
      return
    }
    contructor.value = arg.constructor.name
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
      return
    }
  }
  if (typeof arg === 'string') {
    if (props.depth == 0) {
      typ.value = 'raw'
      text.value = arg
      return
    }
    text.value = `'${arg}'`
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
  <span>
    <span :class="{ 'cursor-pointer': entries && expandable }" @click="expanded = expandable ? !expanded : expanded">
      <span v-if="id || id === 0">
        <span>
          <TheToggle :isOpen="expanded" :class="{ invisible: !expandable || !entries }" />
        </span>
        <span class="text-blue-400">{{ id }}:&nbsp;</span>
      </span>
      <span v-else>
        <span v-if="!minify && entries">
          <TheToggle :isOpen="expanded" />
        </span>
      </span>
      <span :class="typ">
        <template v-if="entries">
          <ArgEntries :typ="typ" :entries="entries" :depth="depth" :contructor="contructor" :minify="minify"
            :expanded="expanded" />
        </template>
        <template v-else-if="typ === 'function'">
          <ArgFunction :arg="arg" :minify="minify" />
        </template>
        <template v-else>
          <span>{{ text }}</span>
        </template>
      </span>
    </span>
    <div v-if="entries && expanded">
      <ListEntries :typ="typ" :entries="entries" :depth="depth" :minify="minify" :expanded="expanded" />
    </div>
  </span>
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
