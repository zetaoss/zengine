<!-- AvatarIcon.vue -->
<script setup lang="ts">
import { computed, type PropType } from 'vue'
import type { Avatar } from './avatar'
import IconLetter from './IconLetter.vue'
import IconIdenticon from './IconIdenticon.vue'

type AvatarType = 1 | 2 | 3

const props = defineProps({
  avatar: { type: Object as PropType<Avatar | null>, default: null },
  size: { type: Number, default: 24 },
  showBorder: { type: Boolean, default: false },
  tempType: { type: Number as PropType<AvatarType | null>, default: null },
})

function isAvatarType(v: unknown): v is AvatarType {
  return v === 1 || v === 2 || v === 3
}

const effectiveType = computed<AvatarType | null>(() => {
  if (!props.avatar) return null
  return isAvatarType(props.tempType) ? props.tempType : (isAvatarType(props.avatar.t) ? props.avatar.t : null)
})

</script>

<template>
  <span
    class="inline-flex items-center justify-center overflow-hidden rounded-full box-border align-middle relative hover:z-40 hover:scale-125"
    :class="{
      'ring-2 ring-white dark:ring-gray-900 outline outline-1 -outline-offset-1 outline-black/5 dark:outline-white/10': showBorder
    }" :style="{ height: `${size}px`, width: `${size}px`, background: '#f0f0f0' }" :title="avatar?.name ?? ''">
    <template v-if="!avatar">
      <IconLetter name="?" :size="size" />
    </template>
    <img v-else-if="effectiveType === 3" class="w-full h-full"
      :src="`//www.gravatar.com/avatar/${avatar.ghash}?s=${size}`" />
    <IconLetter v-else-if="effectiveType === 2" :name="avatar.name" :size="size" />
    <IconIdenticon v-else :name="avatar.name" :size="size" />
  </span>
</template>
