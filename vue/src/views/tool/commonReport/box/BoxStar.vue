<script setup lang="ts">
import { mdiStar } from '@mdi/js'

import TheIcon from '@common/components/TheIcon.vue'

defineProps({
  n: { type: Number, required: true },
})
</script>

<template>
  <TheIcon v-for="i in n" :key="i" style="margin:-1px" :size="12" :path="mdiStar" />
</template>
