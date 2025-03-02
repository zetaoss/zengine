<script setup lang="ts">
import { type Item } from './utils';
import BriefCollection from './BriefCollection.vue';
import BriefScalar from './BriefScalar.vue';

withDefaults(defineProps<{
  item: Item;
  expandableIfCollection?: boolean;
}>(), {
  expandableIfCollection: false
});
</script>

<template>
  <template v-if="item.items && item.items.length > 0">
    <BriefCollection :item="item" :expandableIfCollection="expandableIfCollection" />
  </template>
  <template v-else>
    <BriefScalar :item="item" />
  </template>
</template>
