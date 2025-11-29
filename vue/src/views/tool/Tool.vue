<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <div class="h-full grid md:grid-cols-[120px_1fr]">
    <aside class="h-full p-4 z-bg-muted">
      <p class="text-lg font-bold mb-3">도구</p>

      <RouterLink v-for="item in navItems" :key="item.to" :to="item.to" custom v-slot="{ navigate, isActive }">
        <button type="button" @click="navigate"
          class="block w-full text-left px-3 py-2 rounded transition hover:bg-gray-200 dark:hover:bg-gray-800"
          :class="isActive ? 'bg-gray-200 dark:bg-gray-800 font-semibold' : ''">
          {{ item.label }}
        </button>
      </RouterLink>
    </aside>

    <div>
      <RouterView />
    </div>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView } from "vue-router"

interface NavItem {
  to: string
  label: string
}

const navItems: NavItem[] = [
  { to: "/tool/common-report", label: "통용" },
  { to: "/tool/write-request", label: "작성요청" },
  { to: "/tool/frontplay", label: "FrontPlay" },
]
</script>
