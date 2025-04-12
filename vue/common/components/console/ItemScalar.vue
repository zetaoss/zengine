<script setup lang="ts">
import { type Item } from './utils';

defineProps<{
  item: Item;
}>();
</script>

<template>
  <template v-if="item.type === 'string'">
    <span class="string">"{{ item.arg }}"</span>
  </template>
  <template v-else-if="item.type === 'number'">
    <span class="number">{{ item.arg }}</span>
  </template>
  <template v-else-if="item.type === 'boolean'">
    <span class="boolean">{{ item.arg }}</span>
  </template>
  <template v-else-if="item.type === 'null'">
    <span class="null">null</span>
  </template>
  <template v-else-if="item.type === 'undefined'">
    <span class="undefined">undefined</span>
  </template>
  <template v-else>
    <span class="raw">{{ item.arg }}</span>
  </template>
</template>
