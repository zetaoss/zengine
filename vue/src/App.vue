<script setup lang="ts">
import { RouterView } from 'vue-router'

import CLayoutFoot from '@common/components/CLayoutFoot.vue'
import CLayoutRemocon from '@common/components/CLayoutRemocon.vue'

import NavbarMain from '@/components/navbar/NavbarMain.vue'
</script>

<template>
  <div class="h-full min-h-screen flex flex-col">
    <NavbarMain />
    <div class="flex-grow w-full max-w-8xl mx-auto">
      <RouterView />
    </div>
    <CLayoutFoot />
  </div>
  <div class="fixed bottom-0 right-0 z-40">
    <CLayoutRemocon />
  </div>
</template>
