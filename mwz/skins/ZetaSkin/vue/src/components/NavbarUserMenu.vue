<!-- NavbarUserMenu.vue -->
<script setup lang="ts">
import AvatarIcon from '@common/components/avatar/AvatarIcon.vue'
import type Item from '@common/components/navbar/types'
import { useDismissable } from '@common/composables/useDismissable'
import ZIcon from '@common/ui/ZIcon.vue'
import { mdiAccount } from '@mdi/js'
import { type PropType, ref } from 'vue'

import getRLCONF from '@/utils/rlconf'

defineProps({
  userMenu: { type: {} as PropType<Item[]>, required: true },
})

const { avatar } = getRLCONF()

const root = ref<HTMLElement | null>(null)
const open = ref(false)

function toggle() {
  open.value = !open.value
}
function close() {
  open.value = false
}

useDismissable(root, {
  enabled: open,
  onDismiss: close,
})
</script>

<template>
  <div ref="root" class="md:group order-2 ml-auto contents md:relative md:inline-block">
    <button type="button" class="order-2 ml-auto flex h-12 cursor-pointer items-center px-3 hover:bg-gray-800"
      :class="{ 'bg-gray-800': open }" :aria-expanded="open" @click="toggle">
      <span v-if="avatar && avatar.id > 0">
        <AvatarIcon :avatar="avatar" />
      </span>
      <span v-else>
        <ZIcon :path="mdiAccount" />
      </span>
    </button>

    <div class="order-3 z-40
         bg-gray-800 md:absolute md:right-0 md:m-1 md:rounded md:group-hover:block
         md:border w-full md:w-auto" :class="open ? 'block' : 'hidden'" @click.stop>
      <nav class="grid grid-cols-3 w-full py-1 md:w-fit md:block md:whitespace-nowrap">
        <a v-for="item in Object.values(userMenu)" :key="item.text" :href="item.href"
          class="block p-2 px-8 text-xs text-white hover:bg-gray-700 hover:no-underline" @click="close">
          {{ item.text }}
        </a>
      </nav>
    </div>
  </div>
</template>
