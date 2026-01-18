<!-- App.vue -->
<script setup lang="ts">
import CSiteFooter from '@common/components/CSiteFooter.vue'
import CSiteRemocon from '@common/components/CSiteRemocon.vue'
import ConfirmHost from '@common/ui/confirm/ConfirmHost.vue'
import ToastHost from '@common/ui/toast/ToastHost.vue'
import { RouterView } from 'vue-router'

import NavbarMain from '@/components/navbar/NavbarMain.vue'
</script>

<template>
  <ConfirmHost />
  <ToastHost />
  <div class="h-full min-h-screen flex flex-col">
    <NavbarMain />
    <main class="flex-1 w-full max-w-8xl mx-auto">
      <RouterView />
    </main>
    <CSiteFooter />
  </div>
  <div class="fixed bottom-0 right-0 z-40">
    <CSiteRemocon />
  </div>
</template>
