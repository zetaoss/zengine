<script setup lang="ts">
import CNavbarSearch from '@common/components/navbar/CNavbarSearch.vue'

import NavbarUserMenu from './NavbarUserMenu.vue'
</script>

<template>
  <nav id="navbar" class="text-sm bg-slate-700 text-white">
    <div class="flex flex-wrap items-center max-w-8xl mx-auto w-full">
      <div class="order-1 flex items-stretch h-12">
        <RouterLink class="navlink" to="/">
          <img alt="zetawiki" src="/zeta.svg" class="h-6 w-6" />
        </RouterLink>
        <RouterLink class="navlink" to="/" :class="{ '!text-yellow-200': $route.meta.tab == 'wiki' }">
          위키
        </RouterLink>
        <RouterLink class="navlink" to="/forum" :class="{ '!text-yellow-200': $route.path.startsWith('/forum') }">
          포럼
        </RouterLink>
        <RouterLink class="navlink" to="/tool/common-report"
          :class="{ '!text-yellow-200': $route.path.startsWith('/tool') }">
          도구
        </RouterLink>
      </div>
      <div class="order-2 contents md:flex md:items-stretch md:ml-auto md:w-auto md:flex-1 md:max-w-2xl">
        <div class="md:group order-2 ml-auto contents md:relative md:inline-block">
          <NavbarUserMenu />
        </div>
        <div class="order-3 basis-full min-w-0 md:order-1 md:basis-auto md:flex-1">
          <div v-if="$route.meta.tab == 'wiki'">
            <CNavbarSearch />
          </div>
        </div>
      </div>
    </div>
  </nav>
</template>
