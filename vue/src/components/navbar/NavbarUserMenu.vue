<!-- NavbarUserMenu.vue -->
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { mdiAccount } from '@mdi/js'

import ZIcon from '@common/ui/ZIcon.vue'
import AvatarIcon from '@common/components/avatar/AvatarIcon.vue'
import useAuthStore from '@/stores/auth'
import { useDismissable } from '@common/composables/useDismissable'

import type Item from '@common/components/navbar/types'

const route = useRoute()
const me = useAuthStore()
me.update()

const items = computed<Item[]>(() =>
  me.isLoggedIn
    ? [
      { text: '사용자 페이지', href: `/user/${me.userData.avatar.name}` },
      { text: '사용자 문서', href: '/wiki/특수:내사용자문서' },
      { text: '사용자 토론', href: '/wiki/특수:내사용자토론' },
      { text: '환경 설정', href: '/wiki/특수:환경설정' },
      { text: '주시문서 목록', href: '/wiki/특수:주시문서목록' },
      { text: '기여', href: '/wiki/특수:내기여' },
      { text: '업로드', href: '#' },
      { text: '특수문서', href: '#' },
      { text: '로그아웃', href: '/logout' },
    ]
    : [
      { text: '토론', href: '/wiki/특수:내사용자토론' },
      { text: '기여', href: '/wiki/특수:내기여' },
      { text: '계정 생성', href: '/wiki/특수:계정만들기' },
      { text: '로그인', href: '/login' },
    ],
)

const root = ref<HTMLElement | null>(null)
const open = ref(false)

const toggle = () => (open.value = !open.value)
const close = () => (open.value = false)

useDismissable(root, {
  enabled: open,
  onDismiss: close,
})

watch(() => route.path, close)
</script>

<template>
  <div ref="root" class="md:group order-2 ml-auto contents md:relative md:inline-block">
    <button type="button"
      class="order-2 ml-auto flex h-12 w-12 items-center justify-center hover:bg-gray-800 md:w-auto md:px-3"
      :class="{ 'bg-gray-800': open }" :aria-expanded="open" @click="toggle">
      <AvatarIcon v-if="me.isLoggedIn" :avatar="me.userData.avatar" />
      <ZIcon v-else :path="mdiAccount" />
    </button>

    <div class="order-3 z-40 bg-gray-800 md:absolute md:right-0 md:m-1 md:rounded md:border w-full md:w-auto"
      :class="open ? 'block' : 'hidden'" @click.stop>
      <nav class="grid grid-cols-3 w-full py-1 md:block md:w-fit md:whitespace-nowrap">
        <a v-for="item in items" :key="item.text" :href="item.href"
          class="block px-8 p-2 text-xs text-white hover:bg-gray-700 hover:no-underline" @click="close">
          {{ item.text }}
        </a>
      </nav>
    </div>
  </div>
</template>
