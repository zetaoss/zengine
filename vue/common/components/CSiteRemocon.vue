<script setup lang="ts">
import { mdiChevronDown, mdiChevronUp, mdiWeatherNight } from '@mdi/js'
import { useDark, useToggle } from '@vueuse/core'

import ZIcon from '@common/ui/ZIcon.vue'
import { scrollToTop, scrollToBottom } from '@common/utils/scroll'

const isDark = useDark()
const toggleDark = useToggle(isDark)
</script>

<template>
  <div class="fixed bottom-0 right-0 z-40 print:hidden">
    <div class="z-50 mx-0.5 text-white opacity-80">
      <button type="button" @click="toggleDark()">
        <ZIcon :class="{ 'text-yellow-500': isDark }" :path="mdiWeatherNight" />
      </button>
      <button type="button" @click="scrollToTop">
        <ZIcon :path="mdiChevronUp" />
      </button>
      <button type="button" @click="scrollToBottom">
        <ZIcon :path="mdiChevronDown" />
      </button>
    </div>
  </div>
</template>

<style scoped>
button {
  @apply p-1.5 ml-0.5 rounded bg-[#8888];
}

svg {
  @apply w-6 h-6;
}
</style>
