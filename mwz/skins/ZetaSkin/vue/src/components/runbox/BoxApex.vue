<!-- BoxApex.vue -->
<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import BoxFront from './BoxFront.vue'
import BoxLang from './BoxLang.vue'
import BoxNotebook from './BoxNotebook.vue'
import BoxZero from './BoxZero.vue'
import type { Job } from './types'

import { useClipboard } from '@vueuse/core'
import { mdiContentCopy, mdiCheck } from '@mdi/js'
import ZIcon from '@common/ui/ZIcon.vue'

const props = defineProps<{
  job: Job
  seq: number
}>()

const componentMap = {
  front: BoxFront,
  lang: BoxLang,
  notebook: BoxNotebook,
  zero: BoxZero,
} as const

const CurrentComponent = computed(() => {
  return componentMap[props.job.type as keyof typeof componentMap] ?? BoxZero
})

type SimpleBox = { lang: string; text: string }

const box = computed<SimpleBox>(() => {
  return (props.job.boxes?.[props.seq] as SimpleBox | undefined) ?? {
    lang: '',
    text: '',
  }
})

const loaded = ref(props.job.phase != null)
watch(
  () => props.job.phase,
  (newPhase) => {
    loaded.value = newPhase != null
  },
  { immediate: true },
)

const { copy, copied } = useClipboard()

const onCopy = () => {
  copy(box.value.text)
}
</script>

<template>
  <div v-if="props.job">
    <div class="mb-1 bg-[var(--code-bg)] p-1 border rounded-lg">
      <component :is="CurrentComponent" :job="props.job" :seq="props.seq">
        <div class="pt-1 px-4">
          <div class="sticky top-0 z-10 h-0">
            <div class="flex justify-end">
              <button class="p-1 rounded text-xs z-muted2 inline-flex items-center space-x-1 cursor-pointer"
                @click="onCopy">
                <template v-if="!copied">
                  <ZIcon :size="14" :path="mdiContentCopy" />
                  <span>Copy</span>
                </template>

                <template v-else>
                  <ZIcon :size="14" :path="mdiCheck" />
                  <span>Copied</span>
                </template>
              </button>
            </div>
          </div>

          <div class="text-xs z-muted2 select-none flex items-center gap-2">
            <span>{{ box.lang }}</span>
            <span v-if="props.job.boxes.length > 1">
              <span v-for="(_, i) in props.job.boxes" :key="i" :class="{ 'opacity-30': i !== props.seq }">
                ●
              </span>
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
