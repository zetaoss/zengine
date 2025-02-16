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
  adjustSizes(isVertical.value ? width : height);
});

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
  if (divider.value) {
    divider.value.style.cursor = isVertical.value ? 'ew-resize' : 'ns-resize';
  }
};

const startResize = () => {
  document.body.style.userSelect = 'none';
  isResizing.value = true;
  if (divider.value) {
    divider.value.style.cursor = isVertical.value ? 'col-resize' : 'row-resize';
  }
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
      :class="['relative z-10 bg-gray-300', isVertical ? 'w-1 h-full cursor-ew-resize' : 'h-1 w-full cursor-ns-resize']"
      @mousedown="startResize">
    </div>
    <div class="overflow-auto flex-1">
      <slot name="second"></slot>
    </div>
  </div>
</template>

<style scoped>
/* Divider에만 1px 경계선 추가 */
.bg-gray-300 {
  background-color: #d1d5db;
  /* Divider 배경색 */
}

/* 수직 Divider */
.w-1 {
  width: 1px;
}

/* 수평 Divider */
.h-1 {
  height: 1px;
}
</style>
