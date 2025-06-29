<script setup lang="ts">
import { type PropType, ref } from 'vue'

import BaseIcon from '@common/ui/BaseIcon.vue'
import AvatarCore from '@common/components/avatar/AvatarCore.vue'
import CNavbarItems from '@common/components/navbar/CNavbarItems.vue'
import type Item from '@common/components/navbar/types'
import { mdiAccount, mdiChevronDown } from '@mdi/js'
import { onClickOutside } from '@vueuse/core'

import getRLCONF from '@/utils/rlconf'

defineProps({
  userMenu: { type: {} as PropType<Item[]>, required: true },
})

const { avatar } = getRLCONF()
const show = ref(false)
const target = ref(null)

onClickOutside(target, () => { show.value = false })
</script>

<template>
  <div ref="target" class="h-12 w-12 float-right flex md:flex-none order-2 md:relative">
    <button type="button" class="w-full hover:bg-zinc-200 dark:hover:bg-zinc-800"
      :class="{ 'bg-zinc-200 dark:bg-zinc-800': show }" @click="show = !show">
      <div class="relative">
        <span v-if="avatar && avatar.id > 0">
          <AvatarCore :user-avatar="avatar" />
        </span>
        <span v-else>
          <BaseIcon :path="mdiAccount" />
        </span>
        <span
          class="bg-slate-200 dark:bg-slate-800 rounded-full absolute inline-flex top-[calc(100%-6px)] left-[calc(100%-20px)]">
          <BaseIcon :path="mdiChevronDown" :size="14" />
        </span>
      </div>
    </button>
    <div v-show="show" class="hidden md:block md:absolute z-30 top-full right-0">
      <div class="rounded-lg overflow-hidden m-1 border bg-z-card">
        <CNavbarItems :items="Object.values(userMenu)" />
      </div>
    </div>
  </div>
  <div class="columns-2 py-1 md:hidden bg-gray-50 dark:bg-gray-950" :class="{ hidden: !show }">
    <CNavbarItems :items="Object.values(userMenu)" />
  </div>
</template>
