<script setup lang="ts">
import type { PropType } from 'vue'

import Icon from '@common/ui/Icon.vue'

import type { ItemData } from './types'

defineProps({
  itemData: { type: Object as PropType<ItemData>, required: true },
})
</script>

<template>
  <button type="button" class="menu-item" :class="{ 'is-active': itemData.isActive ? itemData.isActive() : null }"
    :title="itemData.title" @click="itemData.action">
    <Icon :path="itemData.icon" />
  </button>
</template>
