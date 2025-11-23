<!-- ZFold.vue -->
<template>
  <span class="fold-x">
    <slot />
  </span>
</template>

<style scoped>
.fold-x {
  display: inline-block;
  transform-origin: center;
  animation: fold-x 1s ease-in-out infinite;
}

@keyframes fold-x {

  0%,
  100% {
    transform: scaleX(1);
  }

  50% {
    transform: scaleX(0);
  }
}
</style>
