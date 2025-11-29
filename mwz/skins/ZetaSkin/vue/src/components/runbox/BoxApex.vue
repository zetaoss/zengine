<!-- BoxApex.vue -->
<script setup lang="ts">
import { inject, computed, ref, watch } from 'vue'

import BoxFront from './BoxFront.vue'
import BoxLang from './BoxLang.vue'
import BoxNotebook from './BoxNotebook.vue'
import BoxZero from './BoxZero.vue'
import type { Job } from './types'

import { useClipboard } from '@vueuse/core'
import { mdiContentCopy, mdiCheck } from '@mdi/js'
import ZIcon from '@common/ui/ZIcon.vue'

const job = inject('job') as Job
const seq = inject('seq') as number

const componentMap = {
  front: BoxFront,
  lang: BoxLang,
  notebook: BoxNotebook,
  zero: BoxZero,
}

const CurrentComponent = computed(() => {
  return componentMap[job.type as keyof typeof componentMap] ?? BoxZero
})

const box = computed(() => job.boxes?.[seq] ?? { lang: '', text: '' })

const loaded = ref(false)
watch(
  () => job.phase,
  (newPhase) => {
    loaded.value = newPhase !== null && newPhase !== undefined
  },
  { immediate: true },
)

const { copy, copied } = useClipboard()
</script>

<template>
  <div v-if="job">
    <div class="mb-1 bg-[var(--runbox-bg)] p-1 border rounded-lg">
      <component :is="CurrentComponent" :job="job" :seq="seq">
        <div class="bg-[var(--code-bg)] rounded-lg pt-1 px-4">
          <div class="sticky top-0 z-10 h-0">
            <div class="flex justify-end">
              <button class="p-1 rounded text-xs inline-flex items-center space-x-1 cursor-pointer"
                @click="copy(box.text)">
                <template v-if="!copied">
                  <ZIcon :size="14" :path="mdiContentCopy" />
                  <span>Copy code</span>
                </template>

                <template v-else>
                  <ZIcon :size="14" :path="mdiCheck" />
                  <span>Copied</span>
                </template>
              </button>
            </div>
          </div>
          <div class="text-xs opacity-50 select-none flex items-center gap-2">
            <span>{{ box.lang }}</span>
            <span v-if="job.boxes.length > 1">
              <span v-for="(_, i) in job.boxes" :key="i" :class="{ 'opacity-25': i != seq }">●</span>
            </span>
          </div>

          <div class="py-3">
            <slot />
          </div>
        </div>
      </component>
    </div>
  </div>
</template>
