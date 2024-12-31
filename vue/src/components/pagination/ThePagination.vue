<script setup lang="ts">
import { type PropType, ref, watch } from 'vue'

import { mdiChevronLeft, mdiChevronRight } from '@mdi/js'

import TheIcon from '@common/components/TheIcon.vue'

import type { PaginateData } from './types'

const props = defineProps({
  paginateData: { type: Object as PropType<PaginateData>, required: true },
})
const startPage = ref(1)
const endPage = ref(1)
const prevPage = ref(0)
const nextPage = ref(0)
function load() {
  if (!props.paginateData.current_page) {
    return
  }
  startPage.value = Math.floor((props.paginateData.current_page - 1) / 10) * 10 + 1
  endPage.value = Math.min(startPage.value + 9, props.paginateData.last_page)
  prevPage.value = startPage.value - 1
  nextPage.value = endPage.value + 1
  if (nextPage.value > props.paginateData.last_page) {
    nextPage.value = 0
  }
}

watch(() => props.paginateData, load)
load()
</script>

<template>
  <div v-if="paginateData.path" class="leading-4 py-2">
    <span v-if="prevPage">
      <RouterLink :to="{ path: `${paginateData.path}/${prevPage}` }" class="btn btn-light !px-2">
        <TheIcon :path="mdiChevronLeft" />
        이전
      </RouterLink>
    </span>

    <span v-for="page in Array.from({ length: endPage - startPage + 1 }, (_, i) => startPage + i)" :key="page">
      <RouterLink :to="{ path: `${paginateData.path}/${page}` }" class="btn btn-light !px-3"
        :class="{ disabled: page == paginateData.current_page }">{{ page }}
      </RouterLink>
    </span>

    <span v-if="nextPage">
      <RouterLink :to="{ path: `${paginateData.path}/${nextPage}` }" class="btn btn-light !px-2">
        다음
        <TheIcon :path="mdiChevronRight" />
      </RouterLink>
    </span>
  </div>
</template>
