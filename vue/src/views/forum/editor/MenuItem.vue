<script setup lang="ts">
import ZIcon from '@common/ui/ZIcon.vue'
import type { PropType } from 'vue'

import type { ItemData } from './types'

defineProps({
  itemData: { type: Object as PropType<ItemData>, required: true },
})
</script>

<template>
  <button type="button" class="menu-item" :class="{ 'is-active': itemData.isActive ? itemData.isActive() : null }"
    :title="itemData.title" @click="itemData.action">
    <ZIcon :path="itemData.icon" />
  </button>
</template>
