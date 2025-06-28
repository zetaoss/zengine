<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useClipboard } from '@vueuse/core'
import { type Job } from './types'
import { mdiContentCopy, mdiCheck, mdiLoading } from '@mdi/js'
import Icon from '@common/ui/Icon.vue'

const props = defineProps<{
  job: Job
  seq?: number
}>()

const box = computed(() => {
  const seq = props.seq ?? 0
  return props.job?.boxes?.[seq] ?? { lang: '', text: '' }
})

const boxCount = computed(() => props.job?.boxes?.length ?? 0);

const logs = computed(() => props.job && props.job.main === props.seq ? props.job.logs : []);

const loaded = ref(false);

const { copy, copied } = useClipboard()

watch(() => props.job.phase, (newPhase) => {
  if (newPhase !== null) {
    loaded.value = true;
    console.log("Job phase changed:", props.job);
  }
}, { immediate: true });
</script>

<template>
  <div :class="[
    'relative my-1 bg-zinc-100 dark:bg-zinc-800 group border-x-4',
    seq === 0 ? 'rounded-t-lg' : '',
    seq === boxCount - 1 ? 'rounded-b-lg' : ''
  ]">
    <div class="absolute top-1 right-2 z-10 text-xs opacity-50 flex items-center space-x-2 h-[20px]">

      <span v-if="!copied" class="font-bold flex items-center space-x-1">
        <span>{{ box.lang }}</span>
        <Icon v-if="job?.phase === 'pending' || job?.phase === 'running'" :size="16" :path="mdiLoading" :spin="true" />
      </span>

      <div v-else class="items-center space-x-1 text-green-500 inline-flex">
        <Icon :size="16" :path="mdiCheck" />
        <span>copied</span>
      </div>

      <button v-if="!copied"
        class="p-1 mt-1 rounded bg-[#8882] hover:bg-[#8884] hidden group-hover:inline-flex items-center"
        @click="copy(box.text)">
        <Icon :size="18" :path="mdiContentCopy" />
      </button>
    </div>

    <div class="p-3">
      <slot />
    </div>

    <div v-if="loaded && job">
      <div v-if="job?.phase === 'succeeded'">
        <div v-if="logs.length" class="rounded-b-lg bg-black font-mono p-3 break-all">
          <div v-for="(log, index) in logs" :key="index">
            <div :class="{ 'text-red-500': log.charAt(0) === '2' }">
              {{ log.slice(1) }}
            </div>
          </div>
        </div>
      </div>
    </div>

  </div>
</template>
