<!-- BoxLang.vue -->
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { Job } from './types'

const props = defineProps<{ job: Job; seq?: number }>()

const seq = computed(() => props.seq ?? 0)
const job = computed(() => props.job)

const isMain = computed(() => seq.value === job.value?.main)

const logs = computed(() => job.value?.langOuts?.logs ?? [])
const images = computed(() => job.value?.langOuts?.images ?? [])

const loaded = ref(false)

watch(
  () => job.value?.phase,
  (newPhase) => {
    loaded.value = newPhase !== null && newPhase !== undefined
  },
  { immediate: true },
)
</script>

<template>
  <slot />
  <div v-if="loaded && job && isMain">
    <div v-if="job.phase === 'succeeded'" class="mt-1 rounded-lg p-3 break-all">
      <div v-if="logs.length || images.length" :key="'outputs'">
        <template v-if="logs.length && !(job.boxes[seq].lang.includes('tex') && images.length)">
          <pre v-for="(log, index) in logs"
            :key="index"><span :class="{ 'text-red-500': log?.charAt(0) === '2' }">{{ log?.slice(1) }}</span></pre>
        </template>
        <template v-if="images.length">
          <span v-for="(img, index) in images" :key="index">
            <img :src="'data:image/png;base64,' + img" class="bg-white mb-3 mr-3 max-h-96 max-w-96" />
          </span>
        </template>
      </div>
    </div>
  </div>
</template>
