<!-- AvatarIcon.vue -->
<script setup lang="ts">
import type { PropType } from 'vue'
import type { Avatar } from './avatar'
import IconLetter from './IconLetter.vue'
import IconIdenticon from './IconIdenticon.vue'

defineProps({
  avatar: { type: Object as PropType<Avatar | null>, default: null },
  size: { type: Number, default: 24 },
  showBorder: { type: Boolean, default: false },
})
</script>

<template>
  <span class="inline-flex items-center justify-center overflow-hidden rounded-full box-border align-middle" :class="{
    'ring-2 ring-white dark:ring-gray-900 outline outline-1 -outline-offset-1 outline-black/5 dark:outline-white/10': showBorder
  }" :style="{ height: `${size}px`, width: `${size}px`, background: '#f0f0f0' }">
    <template v-if="!avatar">
      <IconLetter name="?" :size="size" />
    </template>

    <img v-else-if="avatar.t === 3" class="w-full h-full" alt="gravatar"
      :src="`//www.gravatar.com/avatar/${avatar.ghash}?s=${size}`" />

    <IconLetter v-else-if="avatar.t === 2" :name="avatar.name" :size="size" />

    <IconIdenticon v-else :name="avatar.name" :size="size" />
  </span>
</template>
