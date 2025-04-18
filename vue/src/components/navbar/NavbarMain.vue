<script setup lang="ts">
import NavbarSearch from '@common/components/navbar/NavbarSearch.vue'
import NavbarUserMenu from './NavbarUserMenu.vue'
</script>

<template>
  <header id="header" class="w-full text-sm border-b bg-white dark:bg-slate-700">
    <div class="max-w-8xl mx-auto w-full md:flex whitespace-nowrap">
      <div class="inline-block md:flex-none">
        <div>
          <div class="flex h-12">
            <RouterLink class="navbar-link" to="/">
              <img alt="zetawiki" src="/zeta.svg" width="12" height="16">
              <span class="hidden lg:inline ml-2 text-lg font-semibold">제타위키</span>
            </RouterLink>
            <RouterLink class="navbar-link" to="/"
              :class="{ '!text-yellow-600 dark:!text-yellow-200': $route.meta.tab == 'wiki' }">
              위키
            </RouterLink>
            <RouterLink class="navbar-link" to="/forum"
              :class="{ '!text-yellow-600 dark:!text-yellow-200': $route.path.startsWith('/forum') }">
              포럼
            </RouterLink>
            <RouterLink class="navbar-link" to="/tool/common-report"
              :class="{ '!text-yellow-600 dark:!text-yellow-200': $route.path.startsWith('/tool') }">
              도구
            </RouterLink>
          </div>
        </div>
      </div>
      <NavbarUserMenu />
      <div class="w-full">
        <div v-if="$route.meta.tab == 'wiki'">
          <NavbarSearch />
        </div>
      </div>
    </div>
  </header>
</template>
