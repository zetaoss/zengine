<script lang="ts" setup>
import { inject, computed } from "vue";
import { type Job } from "./types";

// Safe inject with fallback
const job = inject<Job | null>("job", null);
const seq = inject<number>("seq", -1);

if (!job || seq === -1) {
  throw new Error("Missing injected 'job' or 'seq'");
}

const isMainJob = computed(() => job && job.main === seq);
const phase = computed(() => isMainJob.value ? job!.phase : null);

const logs = computed(() =>
  isMainJob.value ? job!.outs.logs ?? [] : []
);

const images = computed(() =>
  isMainJob.value ? job!.outs.images ?? [] : []
);

const isErrorLog = (log: string) => log.startsWith("2");
</script>

<template>
  <div class="border p-2">
    <slot />
  </div>

  <div v-if="phase === 'pending' || phase === 'running'" class="my-2 text-center">
    <div class="inline-block animate-spin mr-2">⌛</div>
    <span class="font-medium text-gray-700">{{ phase }}</span>
  </div>

  <div v-if="logs.length" class="border font-mono p-2 my-2">
    <div v-for="(log, index) in logs" :key="`log-${index}-${log.slice(0, 10)}`">
      <div :class="{ 'text-red-500': isErrorLog(log) }">
        {{ log.slice(1) }}
      </div>
    </div>
  </div>

  <div v-if="images.length" class="border p-2 mt-2 space-y-2">
    <div v-for="(img, index) in images" :key="`img-${index}-${img.slice(0, 10)}`">
      <img :src="`data:image/png;base64,${img}`" class="bg-white max-w-full" />
    </div>
  </div>
</template>
