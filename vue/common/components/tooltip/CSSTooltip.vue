<script setup lang="ts">
defineProps({
  tooltipText: { type: String, default: 'hello' },
  borderBottom: { type: Boolean, default: false },
  position: { type: String, default: 'top' },
})
</script>
<template>
  <div class="tooltip" :class="{ 'border-b border-dotted': borderBottom }">
    <slot />
    <span class="tooltiptext" :class="position">{{ tooltipText }}</span>
  </div>
</template>
<style scoped lang="scss">
.tooltip {
  @apply relative inline-block cursor-default;

  &:hover {
    .tooltiptext {
      @apply visible;
    }
  }
}

.tooltiptext {
  @apply invisible px-4 bg-gray-500 text-white text-center py-1 absolute z-10 rounded-md;

  &::after {
    @apply absolute border-transparent border-8;
    content: '';
  }
}

.top {
  @apply bottom-full left-1/2 -translate-x-1/2 mb-2;

  &::after {
    @apply top-full left-1/2 -ml-2 border-t-gray-500;
  }
}

.bottom {
  @apply top-[120%] left-1/2 -translate-x-1/2;

  &::after {
    @apply bottom-full left-1/2 -ml-2 border-b-gray-500;
  }
}

.right {
  @apply top-1/2 left-full -translate-y-1/2 ml-2;

  &::after {
    @apply top-1/2 right-full -mt-2 border-r-gray-500;
  }
}

.left {
  @apply top-1/2 right-full -translate-y-1/2 mr-2;

  &::after {
    @apply top-1/2 left-full -mt-2 border-l-gray-500;
  }
}
</style>
