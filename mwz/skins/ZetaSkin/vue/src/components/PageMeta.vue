<!-- PageMeta.vue -->
<script setup lang="ts">
import AvatarIcon from '@common/components/avatar/AvatarIcon.vue'
import ZIcon from '@common/ui/ZIcon.vue'
import { mdiClockOutline } from '@mdi/js';

import getRLCONF from '@/utils/rlconf'

defineProps({
  historyhref: { type: String, required: true },
})

const { lastmod, contributors } = getRLCONF()
</script>

<template>
  <template v-if="contributors.length">
    <a :href="historyhref" class="z-text2">
      <ZIcon :path="mdiClockOutline" />
      {{ `${lastmod.substring(0, 4)}-${lastmod.substring(4, 6)}-${lastmod.substring(6, 8)}` }}
    </a>
    <span class="pl-3 -space-x-0.5">
      <a v-for="u in contributors" :key="u.id" :href="`/profile/${u.name}`">
        <AvatarIcon :user="u" :showBorder="true" />
      </a>
    </span>
  </template>
</template>
