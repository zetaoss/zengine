<!-- ToastHost.vue -->
<template>
  <Teleport to="body">
    <div class="fixed top-2 right-2 z-50">
      <TransitionGroup name="fade" tag="div">
        <div v-for="t in toasts" :key="t.id" class="w-56 mt-1 rounded bg-slate-500/80 text-white shadow-md">
          <div class="px-4 py-2">{{ t.message }}</div>
          <hr class="toast-bar h-1 bg-black/50 rounded" :style="{ animationDuration: `${t.timeout}ms` }" />
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { toasts } from './toast'
</script>

<style scoped>
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(6px);
}

.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

@keyframes toast-bar {
  from {
    width: 100%;
  }

  to {
    width: 0%;
  }
}

.toast-bar {
  animation: toast-bar linear forwards;
}
</style>
