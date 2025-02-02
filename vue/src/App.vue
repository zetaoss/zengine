<script setup lang="ts">
import { RouterView } from 'vue-router'
import NavbarMain from '@/components/navbar/NavbarMain.vue'
import LayoutFoot from '@common/components/LayoutFoot.vue'
import LayoutRemocon from '@common/components/LayoutRemocon.vue'
</script>

<template>
  <div class="min-h-screen flex flex-col">
    <NavbarMain />
    <div class="max-w-8xl mx-auto w-full grow shrink">
      <RouterView />
    </div>
    <LayoutFoot />
  </div>
  <div class="fixed bottom-0 right-0 z-40">
    <LayoutRemocon />
  </div>
</template>
