<script setup lang="ts">
import { mdiChevronDown, mdiChevronUp, mdiWeatherNight } from '@mdi/js'
import { useDark, useToggle } from '@vueuse/core'

import Icon from '@common/ui/Icon.vue'

const isDark = useDark()
const toggleDark = useToggle(isDark)

function scrollToTop() {
  window.scrollTo({
    top: 0,
    left: 0,
    behavior: 'smooth',
  })
}

function scrollToBottom() {
  window.scrollTo({
    top: document.body.scrollHeight,
    left: 0,
    behavior: 'smooth',
  })
}
</script>

<template>
  <div class="fixed bottom-0 right-0 z-40 print:hidden">
    <div class="z-50 mx-0.5 text-white opacity-80">
      <button type="button" @click="toggleDark()">
        <Icon :class="{ 'text-yellow-500': isDark }" :path="mdiWeatherNight" />
      </button>
      <button type="button" @click="scrollToTop">
        <Icon :path="mdiChevronUp" />
      </button>
      <button type="button" @click="scrollToBottom">
        <Icon :path="mdiChevronDown" />
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
