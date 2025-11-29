<script setup lang="ts">
import AvatarIcon from '@common/components/avatar/AvatarIcon.vue'

import getRLCONF from '@/utils/rlconf'

defineProps({
  historyhref: { type: String, required: true },
})

const { lastmod, contributors } = getRLCONF()
</script>
<template>
  <a :href="historyhref" class="z-muted">
    수정 {{ `${lastmod.substring(0, 4)}-${lastmod.substring(4, 6)}-${lastmod.substring(6, 8)}` }}
  </a>
  <a :href="`/profile/${user.name}`" class="page-btn" v-for="user in contributors" :key="user.id">
    <AvatarIcon :userAvatar="user" />
  </a>
</template>
