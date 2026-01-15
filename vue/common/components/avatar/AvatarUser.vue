<!-- @common/components/avatar/AvatarUser.vue -->
<script setup lang="ts">
import type { User } from '@common/types/user'
import type { PropType } from 'vue'

import AvatarIcon from './AvatarIcon.vue'

defineProps({
  user: { type: Object as PropType<User>, required: true },
  size: { type: Number, default: 18 },
  showName: { type: Boolean, default: true },
  showLink: { type: Boolean, default: true },
  showBorder: { type: Boolean, default: false },
})
</script>

<template>
  <component :is="showLink ? 'a' : 'span'" :href="showLink ? `/user/${user.name}` : undefined"
    :class="['inline-flex items-center align-middle', showLink ? 'rounded-lg text-gray-400 hover:no-underline hover:bg-gray-200 dark:hover:bg-gray-700' : '',]">
    <AvatarIcon :user="user" :size="size" :showBorder="showBorder" />
    <span v-if="showName" class="px-1">{{ user.name }}</span>
  </component>
</template>
