<script setup lang="ts">
import type { PropType } from 'vue'

import BaseIcon from '@common/ui/BaseIcon.vue'

import type { ItemData } from './types'

defineProps({
  itemData: { type: Object as PropType<ItemData>, required: true },
})
</script>

<template>
  <button type="button" class="menu-item" :class="{ 'is-active': itemData.isActive ? itemData.isActive() : null }"
    :title="itemData.title" @click="itemData.action">
    <BaseIcon :path="itemData.icon" />
  </button>
</template>
