<!-- @common/components/avatar/AvatarIcon.vue -->
<script setup lang="ts">
import { getAvatarBaseUrl } from '@common/config'
import type { User } from '@common/types/user'
import { computed, type PropType } from 'vue'

type AvatarType = 0 | 1 | 2 | 3

const props = defineProps({
  user: { type: Object as PropType<User>, default: null },
  size: { type: Number, default: 24 },
  typ: { type: Number as PropType<AvatarType>, default: 0 },
  showBorder: { type: Boolean, default: false },
})

const baseUrl = getAvatarBaseUrl()

const src = computed(() => {
  if (!props.user) return ''

  const u = new URL(`${baseUrl}/u/${props.user.id}`)
  u.searchParams.set('s', String(props.size))
  if (props.typ) u.searchParams.set('t', String(props.typ))

  const v = localStorage.getItem('v')
  if (v) u.searchParams.set('v', v)

  return u.toString()
})
</script>

<template>
  <span
    class="inline-flex items-center justify-center overflow-hidden rounded-full box-border align-middle relative hover:z-40"
    :class="{ 'ring-2 ring-white dark:ring-gray-900 outline -outline-offset-1 outline-black/5 dark:outline-white/10': showBorder }"
    :style="{ height: `${size}px`, width: `${size}px`, background: '#f0f0f0' }" :title="user.name">
    <span v-if="user.id == 0" class="w-full h-full flex items-center justify-center text-xs text-gray-500">?</span>
    <img v-else class="w-full h-full" :src="src" :width="size" :height="size" loading="lazy" decoding="async"
      referrerpolicy="no-referrer" />
  </span>
</template>
