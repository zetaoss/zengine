<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useResizeObserver, useEventListener } from '@vueuse/core';

const props = defineProps<{
  direction: 'horizontal' | 'vertical';
  initialPercentage?: number;
}>();

const container = ref<HTMLElement | null>(null);
const divider = ref<HTMLElement | null>(null);
const firstPanePercentage = ref(props.initialPercentage ?? 50);
const isResizing = ref(false);
const MIN_SIZE = 100;

const isVertical = computed(() => props.direction === 'vertical');

const adjustSizes = (containerSize: number) => {
  const minPercentage = (MIN_SIZE / containerSize) * 100;
  const maxPercentage = 100 - minPercentage;
  firstPanePercentage.value = Math.max(minPercentage, Math.min(maxPercentage, firstPanePercentage.value));
};

useResizeObserver(container, (entries) => {
  const { width, height } = entries[0].contentRect;
  adjustSizes(isVertical.value ? height : width);
});

const disableIframes = () => {
  document.querySelectorAll('iframe').forEach((iframe) => {
    iframe.style.pointerEvents = 'none';
  });
};

const enableIframes = () => {
  document.querySelectorAll('iframe').forEach((iframe) => {
    iframe.style.pointerEvents = 'auto';
  });
};

const handleMouseMove = (e: MouseEvent) => {
  if (!isResizing.value || !container.value) return;
  const rect = container.value.getBoundingClientRect();
  const containerSize = isVertical.value ? rect.width : rect.height;
  const mousePosition = isVertical.value ? e.clientX - rect.left : e.clientY - rect.top;
  const newPercentage = (mousePosition / containerSize) * 100;

  const minPercentage = (MIN_SIZE / containerSize) * 100;
  const maxPercentage = 100 - minPercentage;
  firstPanePercentage.value = Math.max(minPercentage, Math.min(maxPercentage, newPercentage));
};

const stopResize = () => {
  document.body.style.userSelect = '';
  isResizing.value = false;
  enableIframes();
};

const startResize = () => {
  document.body.style.userSelect = 'none';
  isResizing.value = true;
  disableIframes();
};

useEventListener(window, 'mousemove', handleMouseMove);
useEventListener(window, 'mouseup', stopResize);

onMounted(() => {
  if (container.value) {
    firstPanePercentage.value = props.initialPercentage ?? 50;
  }
});
</script>

<template>
  <div ref="container" :class="['relative w-full h-full flex', isVertical ? 'flex-row' : 'flex-col']">
    <div class="overflow-auto" :style="{ flex: `0 0 ${firstPanePercentage}%` }">
      <slot name="first"></slot>
    </div>
    <div ref="divider"
      :class="['relative z-10 bg-gray-200 dark:bg-gray-800 divider', isVertical ? 'w-2 h-full cursor-ew-resize' : 'h-2 w-full cursor-ns-resize']"
      @mousedown="startResize">
    </div>
    <div class="overflow-auto flex-1">
      <slot name="second"></slot>
    </div>
  </div>
</template>
