<script setup lang="ts">
import type { PropType } from 'vue'

import type UserAvatar from '@common/types/userAvatar'

import AvatarUser from './AvatarUser.vue'

defineProps({
  userAvatar: { type: Object as PropType<UserAvatar>, required: true },
  size: { type: Number, default: 20 },
})
</script>

<template>
  <a class="avatar pt-[2px] px-[5px] pb-[4px] rounded-lg hover:no-underline hover:bg-gray-200 dark:hover:bg-gray-700"
    :href="`/wiki/User:${userAvatar?.name}`">
    <AvatarUser :user-avatar="userAvatar" :size="size" />
  </a>
</template>
