<script setup lang="ts">
import type { PropType } from 'vue'

import type UserAvatar from '@common/types/userAvatar'

import LetterAvatar from './LetterAvatar.vue'
import TheIdenticon from './TheIdenticon.vue'

defineProps({
  userAvatar: { type: Object as PropType<UserAvatar>, required: true },
  size: { type: Number, default: 18 },
})
</script>

<template>
  <span class="inline-block align-middle">
    <span v-if="userAvatar" class="flex rounded-[30%] overflow-hidden" :style="`height:${size}px; width:${size}px`">
      <template v-if="userAvatar.t == 3">
        <img class="w-full h-full" alt="gravatar" :src="`//www.gravatar.com/avatar/${userAvatar?.ghash}?s=32`">
      </template>
      <template v-else-if="userAvatar.t == 2">
        <LetterAvatar :name="userAvatar.name" :size="size" />
      </template>
      <template v-else>
        <TheIdenticon :name="userAvatar.name" :size="size" />
      </template>
    </span>
  </span>
</template>
