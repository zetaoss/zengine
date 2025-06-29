<!-- BoxLang.vue -->
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useClipboard } from '@vueuse/core'
import { type Job } from './types'
import { mdiContentCopy, mdiCheck, mdiLoading } from '@mdi/js'
import Icon from '@common/ui/Icon.vue'

const props = defineProps<{ job: Job; seq?: number }>();

const seq = computed(() => props.seq ?? 0);
const job = computed(() => props.job);
const isMain = computed(() => seq.value === job.value?.main);
const isJobInProgress = computed(() => job.value?.phase === 'pending' || job.value?.phase === 'running');
const box = computed(() => job.value?.boxes?.[seq.value] ?? { lang: '', text: '' });
const boxCount = computed(() => job.value?.boxes?.length ?? 0);

const logs = computed(() => job.value.langOuts?.logs ?? []);
const images = computed(() => job.value.langOuts?.images ?? []);

const loaded = ref(false);

const { copy, copied } = useClipboard();

watch(() => job.value?.phase, (newPhase) => {
  loaded.value = newPhase !== null && newPhase !== undefined;

}, { immediate: true });
</script>

<template>
  <div
    :class="['relative my-1 bg-zinc-100 dark:bg-zinc-800 group border-x-4', seq === 0 ? 'rounded-t-lg' : '', seq === boxCount - 1 ? 'rounded-b-lg' : '']">
    <div class="absolute top-1 right-2 z-10 text-xs opacity-50 flex items-center space-x-2 h-[20px]">

      <span v-if="!copied" class="font-bold flex items-center space-x-1">
        <span>{{ box.lang }}</span>
        <Icon v-if="isJobInProgress" :size="16" :path="mdiLoading" :spin="true" />
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

    <div v-if="loaded && job && isMain">
      <div v-if="job.phase === 'succeeded'" class="rounded-b-lg bg-black font-mono p-3 break-all">
        <div v-if="logs.length || images.length" :key="'outputs'">
          <template v-if="logs.length && !(box.lang.includes('tex') && images.length)">
            <pre v-for="(log, index) in logs"
              :key="index"><span :class="{ 'text-red-500': log?.charAt(0) === '2' }">{{ log?.slice(1) }}</span></pre>
          </template>
          <template v-if="images.length">
            <span v-for="(image, index) in images" :key="index">
              <img :src="'data:image/png;base64,' + image" class="bg-white mb-3 mr-3 max-h-96 max-w-96" />
            </span>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>
