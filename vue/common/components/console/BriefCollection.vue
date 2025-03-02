<script setup lang="ts">
import { ref } from 'vue';
import TheToggle from './TheToggle.vue';
import BriefAny from './BriefAny.vue';
import ItemAny from './ItemAny.vue';
import { type Item } from './utils';

defineProps<{
  item: Item
  expandableIfCollection: boolean
}>();

const expanded = ref(false);
</script>

<template>
  <span :class="{ 'float-left flex flex-col': item.depth == 0 }">
    <template v-if="expandableIfCollection">
      <span class="flex items-center hover:bg-gray-700 cursor-pointer" @click="expanded = !expanded">
        <span class="w-5">
          <TheToggle :isOpen="expanded" />
        </span>
        <span class="flex-1">
          <template v-if="item.type == 'Array'">
            <span>[</span>
            <span v-for="(x, i) in item.items" :key="i">
              <span v-if="i != 0">, </span>
              <BriefAny :item="x" />
            </span>
            <span>]</span>
          </template>
          <template v-else-if="item.type == 'Object'">
            <span>{</span>
            <span v-for="(item, i) in item.items.slice(0, 5)" :key="i">
              <span v-if="i != 0">, </span>
              <span class="graykey">{{ item.key }}</span>
              <span>:&nbsp;</span>
              <BriefAny :item="item" />
            </span>
            <span v-if="item.items.length > 5">, …</span>
            <span>}</span>
          </template>
        </span>
      </span>
    </template>
    <template v-else>
      <template v-if="item.type == 'Array'">
        <span>[</span>
        <span v-for="(x, i) in item.items" :key="i">
          <span v-if="i != 0">, </span>
          <BriefAny :item="x" />
        </span>
        <span>]</span>
      </template>
      <template v-else-if="item.type == 'Object'">
        <span>{</span>
        <span v-for="(item, i) in item.items.slice(0, 5)" :key="i">
          <span v-if="i != 0">, </span>
          <span class="graykey">{{ item.key }}</span>
          <span>:&nbsp;</span>
          <BriefAny :item="item" />
        </span>
        <span v-if="item.items.length > 5">, …</span>
        <span>}</span>
      </template>
    </template>
    <template v-if="expanded">
      <div v-for="(item, i) in item.items" :key="i" class="ml-8">
        <span class="bluekey">{{ item.key }}</span>
        <span>:&nbsp;</span>
        <span>
          <ItemAny :item="item" :expandableIfCollection="true" />
        </span>
      </div>
    </template>
  </span>
</template>
