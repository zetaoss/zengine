<script setup lang="ts">
import { type Item } from './utils';
import ScalarString from './ScalarString.vue';

defineProps<{ item: Item }>();
</script>

<template>
  <template v-if="item.type == 'string'">
    <ScalarString :text="String(item.arg)" />
  </template>
  <template v-else>
    <span :class="item.type">{{ item.text }}</span>
  </template>
</template>
