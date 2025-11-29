<!-- AvatarIcon.vue -->
<script setup lang="ts">
import type { PropType } from 'vue'
import type { Avatar } from './avatar'
import AvatarLetter from './AvatarLetter.vue'
import AvatarIdenticon from './AvatarIdenticon.vue'

defineProps({
  avatar: { type: Object as PropType<Avatar>, required: true },
  size: { type: Number, default: 24 },
  showBorder: { type: Boolean, default: false },
})
</script>

<template>
  <span class="inline-flex items-center justify-center align-middle overflow-hidden rounded-full box-border"
    :class="{ 'ring-2 ring-white dark:ring-gray-900 outline -outline-offset-1 outline-black/5 dark:outline-white/10': showBorder }"
    :style="{ height: `${size}px`, width: `${size}px` }">
    <img v-if="avatar.t === 3" class="w-full h-full" alt="gravatar"
      :src="`//www.gravatar.com/avatar/${avatar.ghash}?s=${size}`" />
    <AvatarLetter v-else-if="avatar.t === 2" :name="avatar.name" :size="size" />
    <AvatarIdenticon v-else :name="avatar.name" :size="size" />
  </span>
</template>
