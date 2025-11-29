<!-- AvatarIcon.vue -->
<script setup lang="ts">
import type { PropType } from 'vue'
import type { Avatar } from './avatar'
import AvatarLetter from './AvatarLetter.vue'
import AvatarIdenticon from './AvatarIdenticon.vue'

defineProps({
  userAvatar: { type: Object as PropType<Avatar>, required: true },
  size: { type: Number, default: 18 },
  showBorder: { type: Boolean, default: false },
})
</script>

<template>
  <span class="inline-flex align-middle items-center justify-center overflow-hidden rounded-[30%] box-border"
    :class="{ border: showBorder }" :style="{ height: `${size}px`, width: `${size}px` }">
    <template v-if="userAvatar.t == 3">
      <img class="w-full h-full" alt="gravatar" :src="`//www.gravatar.com/avatar/${userAvatar.ghash}?s=32`">
    </template>
    <template v-else-if="userAvatar.t == 2">
      <AvatarLetter :name="userAvatar.name" :size="size" />
    </template>
    <template v-else>
      <AvatarIdenticon :name="userAvatar.name" :size="size" />
    </template>
  </span>
</template>
