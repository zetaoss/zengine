<!-- @common/components/avatar/AvatarUser.vue -->
<script setup lang="ts">
import type { PropType } from 'vue'
import type UserAvatar from '@common/types/userAvatar'
import AvatarCore from './AvatarCore.vue'

const props = defineProps({
  userAvatar: { type: Object as PropType<UserAvatar>, required: true },
  size: { type: Number, default: 18 },
  showName: { type: Boolean, default: true },
  showLink: { type: Boolean, default: true },
  showBorder: { type: Boolean, default: true },
})

const tag = props.showLink ? 'a' : 'span'
const href = props.showLink ? `/wiki/User:${props.userAvatar.name}` : undefined
const classVar = [
  'inline-flex items-center align-middle',
  props.showLink ? 'rounded-lg text-gray-400 hover:no-underline hover:bg-gray-200 dark:hover:bg-gray-700' : '',
]
</script>

<template>
  <component :is="tag" :href="href" :class="classVar">
    <AvatarCore :user-avatar="userAvatar" :size="size" :showBorder="showBorder" />
    <span v-if="showName" class="px-1">{{ userAvatar.name }}</span>
  </component>
</template>
