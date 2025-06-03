<!-- eslint-disable vue/multi-word-component-names -->
<script setup lang="ts">
defineProps({
  path: { type: String, required: true },
  size: { type: Number, default: 16 }
})
</script>

<template>
  <svg :width="size" :height="size" viewBox="0 0 24 24" fill="currentColor" xmlns="http://www.w3.org/2000/svg"
    aria-hidden="true" class="inline-block align-middle">
    <path :d="path" />
  </svg>
</template>
