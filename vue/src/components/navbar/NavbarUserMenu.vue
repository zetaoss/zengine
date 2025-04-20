<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { mdiAccount, mdiChevronDown } from '@mdi/js'
import { onClickOutside } from '@vueuse/core'
import { useRoute } from 'vue-router'

import TheIcon from '@common/components/TheIcon.vue'
import AvatarCore from '@common/components/avatar/AvatarCore.vue'
import CNavbarItems from '@common/components/navbar/CNavbarItems.vue'
import useAuthStore from '@/stores/auth'

import getUserMenuItems from './utils'

const route = useRoute()
const me = useAuthStore()
const show = ref(false)
const inside = ref(null)

const items = computed(() => getUserMenuItems(me.userData.avatar))

function close() {
  show.value = false
}

me.update()
onClickOutside(inside, close)
watch(() => route.path, close)
</script>
<template>
  <div ref="inside" class="h-12 w-12 float-right flex md:flex-none order-2 md:relative">
    <button type="button" class="w-full hover:bg-zinc-200 dark:hover:bg-zinc-800"
      :class="{ 'bg-zinc-200 dark:bg-zinc-800': show }" @click="show = !show">
      <div class="relative">
        <span v-if="me.isLoggedIn">
          <AvatarCore :user-avatar="me.userData.avatar" :size="24" />
        </span>
        <span v-else>
          <TheIcon :path="mdiAccount" />
        </span>
        <span
          class="bg-slate-200 dark:bg-slate-800 rounded-full absolute inline-flex top-[calc(100%-6px)] left-[calc(100%-20px)]">
          <TheIcon :path="mdiChevronDown" :size="14" />
        </span>
      </div>
    </button>
    <div v-show="show" class="hidden md:block md:absolute z-30 top-full right-0">
      <div class="rounded-lg overflow-hidden m-1 border bg-z-card">
        <CNavbarItems :items="items" />
      </div>
    </div>
  </div>
  <div v-show="show" class="columns-2 py-1 md:hidden bg-zinc-200 dark:bg-zinc-800">
    <CNavbarItems :items="items" />
  </div>
</template>
