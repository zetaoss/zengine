<script setup lang="ts">
defineProps({
  clss: { type: String, default: '' },
  to: { type: String, required: true },
})
</script>
<template>
  <a :href="to" :class="clss">
    <slot />
  </a>
</template>
