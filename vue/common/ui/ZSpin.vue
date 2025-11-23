<script setup lang="ts">
defineProps({
  reversed: Boolean,
})
</script>

<template>
  <span class="inline-block" :class="reversed ? 'animate-spin-ccw' : 'animate-spin'">
    <slot />
  </span>
</template>

<style scoped>
@keyframes z-spin-ccw {
  0% {
    transform: rotate(360deg);
  }

  100% {
    transform: rotate(0deg);
  }
}

.animate-spin-ccw {
  animation: z-spin-ccw 1s linear infinite;
}
</style>
