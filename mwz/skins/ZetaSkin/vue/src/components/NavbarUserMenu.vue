<script setup lang="ts">
import { type PropType, ref } from 'vue'
import ZIcon from '@common/ui/ZIcon.vue'
import AvatarIcon from '@common/components/avatar/AvatarIcon.vue'
import type Item from '@common/components/navbar/types'
import { mdiAccount } from '@mdi/js'
import { onClickOutside } from '@vueuse/core'
import getRLCONF from '@/utils/rlconf'

defineProps({
  userMenu: { type: {} as PropType<Item[]>, required: true },
})

const { avatar } = getRLCONF()

const cb = ref<HTMLInputElement | null>(null)
const dd = ref<HTMLElement | null>(null)
const lb = ref<HTMLElement | null>(null)

const close = () => {
  const el = cb.value
  if (el && el.checked) el.checked = false
}

onClickOutside(cb, close, { ignore: [lb] })
onClickOutside(dd, close, { ignore: [lb] })
</script>

<template>
  <input id="usermenu-toggle" type="checkbox" ref="cb" class="peer sr-only" />

  <label ref="lb" for="usermenu-toggle"
    class="order-2 ml-auto flex h-12 cursor-pointer items-center px-3 hover:bg-gray-800 peer-checked:bg-gray-800">
    <span v-if="avatar && avatar.id > 0">
      <AvatarIcon :user-avatar="avatar" />
    </span>
    <span v-else>
      <ZIcon :path="mdiAccount" />
    </span>
    <svg class="h-4 w-4 opacity-80" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
      <path d="M7,10L12,15L17,10H7Z" />
    </svg>
  </label>

  <div ref="dd" class="order-3 hidden peer-checked:block z-40
         bg-gray-800 md:absolute md:right-0 md:m-1 md:rounded md:group-hover:block
         md:border w-full md:w-auto">
    <nav class="grid grid-cols-3 w-full py-1 md:w-fit md:block md:whitespace-nowrap">
      <a v-for="item in Object.values(userMenu)" :key="item.text" :href="item.href"
        class="block p-2 px-8 text-xs text-white hover:bg-gray-700 hover:no-underline">
        {{ item.text }}
      </a>
    </nav>
  </div>
</template>
