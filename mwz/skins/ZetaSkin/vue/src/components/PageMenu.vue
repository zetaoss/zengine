<script setup lang="ts">
import { ref } from 'vue'

import Icon from '@common/ui/Icon.vue'
import { mdiDotsVertical } from '@mdi/js'
import { vOnClickOutside } from '@vueuse/components'

const show = ref(false)
</script>

<template>
  <div class="page-menu relative print:hidden" v-on-click-outside="() => { show = false }">
    <button type="button" class="page-btn" @click="show = !show">
      <Icon :path="mdiDotsVertical" :size="16" />
    </button>
    <div v-if="show"
      class="absolute z-10 right-0 p-0 py-2 border list-none rounded bg-white shadow-md dark:bg-black text-black dark:text-white">
      <slot />
    </div>
  </div>
</template>
