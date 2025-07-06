<!-- File: mwz/skins/ZetaSkin/vue/src/components/runbox/BoxNotebook.vue -->
<script setup lang="ts">
import { inject, computed, ref, watch } from 'vue';
import { useClipboard } from '@vueuse/core';
import { mdiContentCopy, mdiCheck, mdiLoading } from '@mdi/js';
import BaseIcon from '@common/ui/BaseIcon.vue';
import NBOutput from './notebook/NBOutput.vue';
import { type Job } from './types';

const job = inject('job') as Job;
const seq = inject('seq') as number;

const nbouts = computed(() => job?.notebookOuts?.[seq] ?? []);
const box = computed(() => job?.boxes?.[seq] ?? { lang: '', text: '' });
const isJobInProgress = computed(() => job?.phase === 'pending' || job?.phase === 'running');

const { copy, copied } = useClipboard();
const loaded = ref(false);

watch(() => job?.phase, (newPhase) => {
  loaded.value = newPhase !== null && newPhase !== undefined;
}, { immediate: true });
</script>

<template>
  <div class="relative border p-2 bg-zinc-100 dark:bg-zinc-800 group">
    <div class="absolute top-1 right-2 z-10 text-xs opacity-50 flex items-center space-x-2 h-[20px]">

      <span v-if="!copied" class="font-bold flex items-center space-x-1">
        <span>{{ box.lang }}</span>
        <BaseIcon v-if="isJobInProgress" :size="16" :path="mdiLoading" :spin="true" />
      </span>

      <div v-else class="items-center space-x-1 text-green-500 inline-flex">
        <BaseIcon :size="16" :path="mdiCheck" />
        <span>copied</span>
      </div>

      <button v-if="!copied"
        class="p-1 mt-1 rounded bg-[#8882] hover:bg-[#8884] hidden group-hover:inline-flex items-center"
        @click="copy(box.text)">
        <BaseIcon :size="18" :path="mdiContentCopy" />
      </button>
    </div>

    <div class="p-3">
      <slot />
    </div>

    <div v-if="loaded && job">
      <div v-if="job.phase === 'succeeded'" class="outputs">
        <div v-if="nbouts && nbouts.length">
          <div v-for="(nbout, i) in nbouts" :key="i">
            <NBOutput :out="nbout" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
