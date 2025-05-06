<script lang="ts" setup>
import { inject, computed, ref, watch } from "vue";
import { type Job } from "./types";

const job = inject<Job>("job");
const seq = inject("seq");

const logs = computed(() => job && job.main === seq ? job.logs : []);

// 로딩 완료 여부를 판단하는 플래그
const loaded = ref(false);

watch(() => job?.phase, (newPhase) => {
  if (newPhase !== null) {
    loaded.value = true;
    console.log("Job phase changed:", job);
  }
}, { immediate: true });
</script>

<template>
  <div class="border p-2">
    <slot />

    <xmp>{{ job }}</xmp>
    <!-- job 데이터 로딩이 완료되었을 때만 표시 -->
    <div v-if="loaded && job" class="mt-2">
      <div v-if="job?.phase === 'succeeded'">
        <div v-if="logs.length" class="border font-mono p-2">
          <div v-for="(log, index) in logs" :key="index">
            <div :class="{ 'text-red-500': log.charAt(0) === '2' }">
              {{ log.slice(1) }}
            </div>
          </div>
        </div>
        <div v-else class="text-gray-400 italic">출력된 로그가 없습니다.</div>
      </div>
      <div v-else class="w-full h-2 bg-gray-200 rounded overflow-hidden mb-2">
        <div class="h-full transition-all duration-300" :class="{
          'bg-gray-400': job.phase === 'pending',
          'bg-blue-500 animate-pulse': job.phase === 'running',
          'bg-red-500': job.phase === 'failed'
        }" />
      </div>
    </div>
  </div>
</template>
