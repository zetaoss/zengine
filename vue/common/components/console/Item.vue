<!-- eslint-disable vue/multi-word-component-names -->
<script setup lang="ts">
import { ref } from 'vue';
import ItemScalar from './ItemScalar.vue';
import Toggle from './Toggle.vue';
import { type Item, inspectArg, stringify } from './utils';

defineProps<{
  item: Item;
}>();

const isOpen = ref(false);
const onClick = () => {
  isOpen.value = !isOpen.value;
};
</script>

<template>
  <template v-if="item.entries">
    <span class="cursor-pointer" @click="onClick">
      <Toggle :isOpen="isOpen" />
      <template v-if="item.type === 'Array'">
        <span>[</span>
        <span v-for="([, v], i) in item.entries" :key="i">
          <span v-if="i != 0">, </span>
          <ItemScalar :item="inspectArg(v)" />
        </span>
        <span>]</span>
      </template>
      <template v-if="item.type === 'Map' || item.type === 'Object'">
        <span>{{ item.name ?? item.type }} {</span>
        <span v-for="([k, v], i) in item.entries.slice(0, 5)" :key="i">
          <span v-if="i != 0">, </span>
          <span class="graykey">{{ k }}</span>
          <span>:&nbsp;</span>
          <ItemScalar :item="inspectArg(v)" />
        </span>
        <span v-if="item.entries.length > 5">, …</span>
        <span>}</span>
      </template>
    </span>
    <div class="pl-3 bg-gradient-to-b from-blue-500 to-pink-500 dark:from-blue-600 dark:to-pink-600" v-if="isOpen">
      <pre class="p-2 bg-slate-100 dark:bg-slate-700" v-text="stringify(item.arg)" />
    </div>
  </template>
  <template v-else>
    <ItemScalar :item="item" />
  </template>
</template>
