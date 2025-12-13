<!-- @common/ui/ZMenuItem.vue -->
<script setup lang="ts">
defineEmits<{ (e: 'click'): void }>()
</script>

<template>
  <button type="button" class="w-full whitespace-nowrap text-left px-3 py-2 hover:bg-[var(--z-btn-hover)]"
    @click="$emit('click')">
    <slot />
  </button>
</template>
