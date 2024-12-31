<script setup lang="ts">
import { ref } from 'vue'

import { useRoute } from 'vue-router'

import AvatarCore from '@common/components/avatar/AvatarCore.vue'
import type UserAvatar from '@common/types/userAvatar'

import UserCalendarHeatmap from './UserCalendarHeatmap.vue'

const userAvatar = ref({
  id: 1,
  name: 'Jmnote',
  ghash: 'asdasdasdasfasf',
  t: 2,
} as UserAvatar)
const route = useRoute()

const name = ref(route.params.name)
</script>

<template>
  <div class="p-3">
    <AvatarCore :size="100" :user-avatar="userAvatar" />
    Profile {{ name }}
    <UserCalendarHeatmap />
  </div>
</template>
