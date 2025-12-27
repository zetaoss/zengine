<!-- ThePagination.vue -->
<script setup lang="ts">
import ZIcon from '@common/ui/ZIcon.vue'
import { mdiChevronLeft, mdiChevronRight } from '@mdi/js'
import { computed, type PropType } from 'vue'
import { RouterLink } from 'vue-router'

interface PaginateMeta {
  current_page: number
  last_page: number
  path: string
}

const props = defineProps({
  paginateData: {
    type: Object as PropType<PaginateMeta>,
    required: true,
  },
  blockSize: {
    type: Number,
    default: 10,
  },
})

const block = computed(() => {
  const { current_page, last_page } = props.paginateData
  const size = props.blockSize

  const blockStart = Math.floor((current_page - 1) / size) * size + 1
  const blockEnd = Math.min(blockStart + size - 1, last_page)

  const pages = Array.from({ length: blockEnd - blockStart + 1 }, (_, i) => blockStart + i)

  return {
    blockStart,
    blockEnd,
    pages,
    hasPrevBlock: blockStart > 1,
    hasNextBlock: blockEnd < last_page,
  }
})

const baseClass =
  'inline-flex items-center justify-center min-w-8 px-2 sm:px-3 py-2 rounded transition z-text ' +
  'hover:no-underline hover:bg-zinc-100 dark:hover:bg-zinc-800 whitespace-nowrap'
const activeClass = 'font-bold bg-[#8883]'
const disabledClass = 'opacity-40 cursor-default pointer-events-none'

const getPageLink = (page: number) => ({
  path: props.paginateData.path,
  query: page > 1 ? { page } : {},
})
</script>

<template>
  <nav class="w-full flex justify-center">
    <div class="max-w-full overflow-x-auto px-2">
      <div class="inline-flex items-center gap-1 whitespace-nowrap">
        <RouterLink :to="getPageLink(block.blockStart - 1)" :class="[baseClass, !block.hasPrevBlock && disabledClass]"
          aria-label="Previous pages">
          <ZIcon :path="mdiChevronLeft" />
        </RouterLink>

        <RouterLink v-for="page in block.pages" :key="page" :to="getPageLink(page)"
          :class="[baseClass, page === paginateData.current_page && activeClass]"
          :aria-current="page === paginateData.current_page ? 'page' : undefined">
          {{ page }}
        </RouterLink>

        <RouterLink :to="getPageLink(block.blockEnd + 1)" :class="[baseClass, !block.hasNextBlock && disabledClass]"
          aria-label="Next pages">
          <ZIcon :path="mdiChevronRight" />
        </RouterLink>
      </div>
    </div>
  </nav>
</template>
