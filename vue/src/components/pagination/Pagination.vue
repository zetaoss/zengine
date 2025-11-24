<!-- eslint-disable vue/multi-word-component-names -->
<script setup lang="ts">
import { computed, type PropType } from 'vue'
import { RouterLink } from 'vue-router'
import { mdiChevronLeft, mdiChevronRight } from '@mdi/js'
import ZIcon from '@common/ui/ZIcon.vue'
import type { PaginateData } from './types'

const props = defineProps({
  paginateData: {
    type: Object as PropType<PaginateData>,
    required: true,
  },
})

const block = computed(() => {
  const { current_page, last_page } = props.paginateData
  const blockSize = 10

  const blockStart = Math.floor((current_page - 1) / blockSize) * blockSize + 1
  const blockEnd = Math.min(blockStart + blockSize - 1, last_page)
  const pages = Array.from({ length: blockEnd - blockStart + 1 }, (_, i) => blockStart + i)

  return { blockStart, blockEnd, pages }
})

const baseLinkClass = 'inline-flex items-center justify-center px-4 py-2 rounded transition z-text hover:no-underline hover:bg-zinc-100 dark:hover:bg-zinc-800'
const disabledClass = 'opacity-40 cursor-default pointer-events-none'
const activeClass = 'font-bold bg-[#8883]'
</script>

<template>
  <nav class="w-full flex justify-center">
    <div v-if="paginateData.path" class="grid mx-auto grid-flow-col">
      <RouterLink :to="{ path: `${paginateData.path}/${block.blockStart - 1}` }"
        :class="[baseLinkClass, block.blockStart < 2 && disabledClass]">
        <ZIcon :path="mdiChevronLeft" />
      </RouterLink>
      <RouterLink v-for="page in block.pages" :key="page" :to="{ path: `${paginateData.path}/${page}` }" :class="[
        baseLinkClass,
        page === paginateData.current_page && activeClass,
      ]">
        {{ page }}
      </RouterLink>
      <RouterLink :to="{ path: `${paginateData.path}/${block.blockEnd + 1}` }"
        :class="[baseLinkClass, block.blockEnd >= paginateData.last_page && disabledClass]">
        <ZIcon :path="mdiChevronRight" />
      </RouterLink>
    </div>
  </nav>
</template>
