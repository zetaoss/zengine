<script setup lang="ts">
defineProps<{ n?: number }>();
</script>

<template>
  <span>{{ '★'.repeat((n ?? 0) + 1) }} {{ ['D', 'C', 'B', 'A', 'S'][n ?? 0] }}</span>
</template>
