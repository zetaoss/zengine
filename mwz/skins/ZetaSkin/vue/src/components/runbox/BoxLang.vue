<!-- BoxLang.vue -->
<script setup lang="ts">
import { computed } from 'vue'

import type { Job } from './types'

const props = defineProps<{
  job: Job
  seq: number
}>()

const { job, seq } = props

const isMain = computed(() => seq === job.main)

const logs = computed(() => job.langOuts?.logs ?? [])
const images = computed(() => job.langOuts?.images ?? [])

const loaded = computed(() => job.phase != null)
const hasLogs = computed(() => logs.value.length > 0)
const hasImages = computed(() => images.value.length > 0)
const hideTexLogs = computed(() => {
  const box = job.boxes[seq]
  return hasLogs.value && hasImages.value && !!box?.lang && box.lang.includes('tex')
})
</script>

<template>
  <slot />

  <div v-if="loaded && isMain">
    <div v-if="job.phase === 'succeeded'" class="rounded-lg bg-[var(--console-bg)] px-4 py-2 text-sm break-all">
      <div v-if="hasLogs || hasImages">
        <template v-if="hasLogs && !hideTexLogs">
          <pre v-for="(log, index) in logs" :key="index" class="whitespace-pre-wrap"
            :class="{ 'text-red-500': log?.charAt(0) === '2' }">{{ log?.slice(1) }}</pre>
        </template>
        <template v-if="hasImages">
          <span v-for="(img, index) in images" :key="index">
            <img :src="`data:image/png;base64,${img}`" class="bg-white mb-3 mr-3 max-h-96 max-w-96" />
          </span>
        </template>
      </div>
    </div>
  </div>
</template>
