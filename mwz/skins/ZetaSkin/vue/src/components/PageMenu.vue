<script setup lang="ts">
import { ref } from 'vue'

import BaseIcon from '@common/ui/BaseIcon.vue'
import { mdiDotsVertical } from '@mdi/js'
import { vOnClickOutside } from '@vueuse/components'

const show = ref(false)
</script>

<template>
  <div class="page-menu relative print:hidden" v-on-click-outside="() => { show = false }">
    <button type="button" class="page-btn" @click="show = !show">
      <BaseIcon :path="mdiDotsVertical" />
    </button>
    <div v-if="show"
      class="absolute z-30 right-0 p-0 py-2 border list-none rounded bg-white shadow-md dark:bg-black text-black dark:text-white">
      <slot />
    </div>
  </div>
</template>
