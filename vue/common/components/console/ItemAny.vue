<script setup lang="ts">
import { computed } from 'vue';
import { getItemCircular, type Item } from './utils';
import BriefCollection from './BriefCollection.vue';
import BriefScalar from './BriefScalar.vue';

const props = withDefaults(defineProps<{
  item: Item;
  expandableIfCollection?: boolean;
}>(), {
  expandableIfCollection: false
});

const myItem = computed(() => {
  if (props.item.type == 'circular') {
    const newItem = getItemCircular(props.item.level, props.item.arg);
    console.log('newItem', newItem);
    return newItem
  }
  return props.item
});
</script>

<template>
  <template v-if="myItem.items && myItem.items.length > 0">
    <BriefCollection :item="myItem" :expandableIfCollection="expandableIfCollection" />
  </template>
  <template v-else>
    <BriefScalar :item="myItem" />
  </template>
</template>
