<script setup lang="ts">
import type { PropType } from 'vue'
import type UserAvatar from '@common/types/userAvatar'
import AvatarCore from './AvatarCore.vue'

defineProps({
  userAvatar: { type: Object as PropType<UserAvatar>, required: true },
  size: { type: Number, default: 18 },
})
</script>

<template>
  <AvatarCore :user-avatar="userAvatar" :size="size" />
  <span class="ml-1">{{ userAvatar?.name }}</span>
</template>
