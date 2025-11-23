<script setup lang="ts">
import { computed } from 'vue'
import { useClipboard } from '@vueuse/core'
import { type Job } from './types'
import { mdiContentCopy, mdiCheck } from '@mdi/js'
import ZIcon from '@common/ui/ZIcon.vue'

const props = defineProps<{
  job: Job
  seq?: number
}>()

const box = computed(() => {
  const seq = props.seq ?? 0
  return props.job?.boxes?.[seq] ?? { lang: '', text: '' }
})

const { copy, copied } = useClipboard()
</script>

<template>
  <div class="relative bg-zinc-100 dark:bg-zinc-800 p-5 rounded group">
    <div class="absolute top-1 right-2 z-10 text-xs opacity-50 flex items-center space-x-2 h-[20px]">

      <span v-if="!copied" class="font-bold group-hover:hidden">
        {{ box.lang }}
      </span>

      <div v-else class="items-center space-x-1 text-green-500 inline-flex">
        <ZIcon :size="16" :path="mdiCheck" />
        <span>copied</span>
      </div>

      <button v-if="!copied"
        class="p-1 mt-1 rounded bg-[#8882] hover:bg-[#8884] hidden group-hover:inline-flex items-center"
        @click="copy(box.text)">
        <ZIcon :size="18" :path="mdiContentCopy" />
      </button>
    </div>

    <div>
      <slot />
    </div>
  </div>
</template>
