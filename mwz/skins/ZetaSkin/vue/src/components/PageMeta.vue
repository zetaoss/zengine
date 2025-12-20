<!-- PageMeta.vue -->
<script setup lang="ts">
import AvatarIcon from '@common/components/avatar/AvatarIcon.vue'
import ZIcon from '@common/ui/ZIcon.vue'
import getRLCONF from '@/utils/rlconf'
import { mdiClockOutline } from '@mdi/js';

defineProps({
  historyhref: { type: String, required: true },
})

const { lastmod, contributors } = getRLCONF()
</script>

<template>
  <template v-if="contributors.length">
    <a :href="historyhref" class="z-muted">
      <ZIcon :path="mdiClockOutline" />
      {{ `${lastmod.substring(0, 4)}-${lastmod.substring(4, 6)}-${lastmod.substring(6, 8)}` }}
    </a>
    <span class="pl-3 -space-x-0.5">
      <a v-for="user in contributors" :key="user.id" :href="`/profile/${user.name}`">
        <AvatarIcon :avatar="user" :showBorder="true" />
      </a>
    </span>
  </template>
</template>
