<script setup lang="ts">
import { defineProps } from 'vue'

defineProps({
  isOpen: Boolean
})
</script>

<template>
  <span class="inline-block text-xs p-1 transition-transform duration-200" :class="isOpen ? 'rotate-90' : 'rotate-0'">
    ▶
  </span>
</template>
