<!-- AvatarIcon.vue -->
<script setup lang="ts">
import type { PropType } from 'vue'
import type { Avatar } from './avatar'
import AvatarLetter from './AvatarLetter.vue'
import AvatarIdenticon from './AvatarIdenticon.vue'

defineProps({
  userAvatar: { type: Object as PropType<Avatar>, required: true },
  size: { type: Number, default: 24 },
  showBorder: { type: Boolean, default: false },
})
</script>

<template>
  <span class="inline-flex items-center justify-center align-middle overflow-hidden rounded-full box-border"
    :class="{ border: showBorder }" :style="{ height: `${size}px`, width: `${size}px` }">
    <img v-if="userAvatar.t === 3" class="w-full h-full" alt="gravatar"
      :src="`//www.gravatar.com/avatar/${userAvatar.ghash}?s=${size}`" />
    <AvatarLetter v-else-if="userAvatar.t === 2" :name="userAvatar.name" :size="size" />
    <AvatarIdenticon v-else :name="userAvatar.name" :size="size" />
  </span>
</template>
