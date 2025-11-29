<!-- UserAvatar.vue -->
<script setup lang="ts">
import type { PropType } from 'vue'
import type { Avatar } from './avatar'
import AvatarIcon from './AvatarIcon.vue'

defineProps({
  userAvatar: { type: Object as PropType<Avatar>, required: true },
  size: { type: Number, default: 18 },
  showName: { type: Boolean, default: true },
  showLink: { type: Boolean, default: true },
  showBorder: { type: Boolean, default: false },
})
</script>

<template>
  <component :is="showLink ? 'a' : 'span'" :href="showLink ? `/wiki/User:${userAvatar.name}` : undefined"
    :class="['inline-flex items-center align-middle', showLink ? 'rounded-lg text-gray-400 hover:no-underline hover:bg-gray-200 dark:hover:bg-gray-700' : '']">
    <AvatarIcon :user-avatar="userAvatar" :size="size" :showBorder="showBorder" />
    <span v-if="showName" class="px-1">{{ userAvatar.name }}</span>
  </component>
</template>
