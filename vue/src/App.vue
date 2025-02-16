<script setup lang="ts">
import { RouterView } from 'vue-router'
import NavbarMain from '@/components/navbar/NavbarMain.vue'
import LayoutFoot from '@common/components/LayoutFoot.vue'
import LayoutRemocon from '@common/components/LayoutRemocon.vue'
</script>

<template>
  <div class="h-full min-h-screen flex flex-col">
    <NavbarMain />
    <div class="flex-grow w-full max-w-8xl mx-auto">
      <RouterView />
    </div>
    <LayoutFoot />
  </div>
  <div class="fixed bottom-0 right-0 z-40">
    <LayoutRemocon />
  </div>
</template>
