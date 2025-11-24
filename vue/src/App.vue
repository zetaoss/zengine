<!-- App.vue -->
<script setup lang="ts">
import { RouterView } from 'vue-router'

import CSiteFooter from '@common/components/CSiteFooter.vue'
import CSiteRemocon from '@common/components/CSiteRemocon.vue'
import GlobalConfirm from '@common/composables/confirm/GlobalConfirm.vue'
import GlobalToast from '@common/composables/toast/GlobalToast.vue'

import NavbarMain from '@/components/navbar/NavbarMain.vue'
</script>

<template>
  <GlobalConfirm />
  <GlobalToast />
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
