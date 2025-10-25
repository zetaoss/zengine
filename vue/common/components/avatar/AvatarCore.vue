<!-- @common/components/avatar/AvatarCore.vue -->
<script setup lang="ts">
import type { PropType } from 'vue'
import type UserAvatar from '@common/types/userAvatar'
import LetterAvatar from './LetterAvatar.vue'
import TheIdenticon from './TheIdenticon.vue'

defineProps({
  userAvatar: { type: Object as PropType<UserAvatar>, required: true },
  size: { type: Number, default: 18 },
  showBorder: { type: Boolean, default: true },
})
</script>

<template>
  <span class="inline-flex align-middle items-center justify-center overflow-hidden rounded-[30%] box-border"
    :class="{ border: showBorder }" :style="{ height: `${size}px`, width: `${size}px` }">
    <template v-if="userAvatar.t == 3">
      <img class="w-full h-full" alt="gravatar" :src="`//www.gravatar.com/avatar/${userAvatar.ghash}?s=32`">
    </template>
    <template v-else-if="userAvatar.t == 2">
      <LetterAvatar :name="userAvatar.name" :size="size" />
    </template>
    <template v-else>
      <TheIdenticon :name="userAvatar.name" :size="size" />
    </template>
  </span>
</template>
