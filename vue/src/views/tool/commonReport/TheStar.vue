<script setup lang="ts">
import { defineProps } from 'vue';

defineProps({ n: Number })
</script>
<template>
  <span>{{ '★'.repeat(n + 1) }} {{ ['D', 'C', 'B', 'A', 'S'][n] }}</span>
</template>
