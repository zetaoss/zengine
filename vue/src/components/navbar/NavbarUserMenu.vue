<!-- NavbarUserMenu.vue -->
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { mdiAccount } from '@mdi/js'

import ZIcon from '@common/ui/ZIcon.vue'
import AvatarIcon from '@common/components/avatar/AvatarIcon.vue'
import useAuthStore from '@/stores/auth'

import { useDismissable } from '@common/composables/useDismissable'
import getUserMenuItems from './utils'

const route = useRoute()
const me = useAuthStore()
me.update()

const items = computed(() => getUserMenuItems(me.userData.avatar))

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

watch(() => route.path, close)
</script>

<template>
  <div ref="root" class="md:group order-2 ml-auto contents md:relative md:inline-block">
    <button type="button"
      class="order-2 ml-auto flex h-12 w-12 cursor-pointer items-center justify-center hover:bg-gray-800 md:w-auto md:px-3"
      :class="{ 'bg-gray-800': open }" :aria-expanded="open" @click="toggle">
      <span v-if="me.isLoggedIn">
        <AvatarIcon :avatar="me.userData.avatar" />
      </span>
      <span v-else>
        <ZIcon :path="mdiAccount" />
      </span>
    </button>
    <div
      class="order-3 z-40 bg-gray-800 md:absolute md:right-0 md:m-1 md:rounded md:group-hover:block md:border w-full md:w-auto"
      :class="open ? 'block' : 'hidden'" @click.stop>
      <nav class="grid grid-cols-3 w-full py-1 md:w-fit md:block md:whitespace-nowrap">
        <a v-for="item in items" :key="item.text" :href="item.href"
          class="block p-2 px-8 text-xs text-white hover:bg-gray-700 hover:no-underline" @click="close">
          {{ item.text }}
        </a>
      </nav>
    </div>
  </div>
</template>
