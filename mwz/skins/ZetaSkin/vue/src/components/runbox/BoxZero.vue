<script setup lang="ts">
import { computed } from 'vue'
import { useClipboard } from '@vueuse/core'
import { type Job } from './types'
import { mdiContentCopy, mdiCheck } from '@mdi/js'
import Icon from '@common/ui/Icon.vue'

const props = defineProps<{
  job: Job
  seq?: number
}>()

const box = computed(() => {
  const seq = props.seq ?? 0
  return props.job?.boxes?.[seq] ?? { lang: '', text: '' }
})

const { copy, copied } = useClipboard({ copiedDuring: 2500 })
</script>

<template>
  <div class="relative">
    <div class="absolute top-1 right-2 z-10 bg-opacity-80 text-xs">
      <div class="flex items-center space-x-2">
        <span>{{ box.lang }}</span>
        <button class="px-2 py-1 rounded bg-gray-800 text-gray-300 hover:text-white transition" @click="copy(box.text)">
          <Icon :path="copied ? mdiCheck : mdiContentCopy" :class="copied ? 'text-green-400' : 'text-gray-300'" />
        </button>
      </div>
    </div>

    <div class="p-2">
      <slot />
    </div>
  </div>
</template>
