<script setup lang="ts">
import TheItem from './TheItem.vue';
import { inspectArg } from './utils';

const props = defineProps<{
  level: string;
  args: unknown[];
}>();

const items = props.args.map((arg) => inspectArg(arg))
</script>

<template>
  <div class="border-b" :class="level">
    <template v-for="(item, i) in items" :key="i">
      <TheItem :item="item" />
    </template>
  </div>
</template>
