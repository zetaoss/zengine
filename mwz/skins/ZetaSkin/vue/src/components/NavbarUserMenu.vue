<script setup lang="ts">
import { type PropType, ref } from 'vue'

import TheIcon from '@common/components/TheIcon.vue'
import AvatarCore from '@common/components/avatar/AvatarCore.vue'
import NavbarItems from '@common/components/navbar/NavbarItems.vue'
import type Item from '@common/components/navbar/types'
import { mdiAccount, mdiChevronDown, mdiChevronUp } from '@mdi/js'
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
    <button type="button" class="w-full hover:bg-slate-200 dark:hover:bg-slate-800" @click="show = !show">
      <span v-if="avatar && avatar.id > 0">
        <AvatarCore :user-avatar="avatar" />
      </span>
      <span v-else>
        <TheIcon :path="mdiAccount" />
      </span>
      <TheIcon :path="show ? mdiChevronUp : mdiChevronDown" />
    </button>
    <div class="hidden md:block md:absolute z-30 top-full right-0" :class="{ 'md:hidden': !show }">
      <div class="rounded border bg-gray-100 dark:bg-gray-900">
        <NavbarItems :items="Object.values(userMenu)" />
      </div>
    </div>
  </div>
  <div class="columns-2 py-1 md:hidden bg-gray-50 dark:bg-gray-950" :class="{ hidden: !show }">
    <NavbarItems :items="Object.values(userMenu)" />
  </div>
</template>
