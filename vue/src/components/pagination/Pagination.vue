<!-- eslint-disable vue/multi-word-component-names -->
<script setup lang="ts">
import { computed, type PropType, watch } from 'vue'
import { mdiChevronLeft, mdiChevronRight } from '@mdi/js'
import BaseIcon from '@common/ui/BaseIcon.vue'
import type { PaginateData } from './types'

const props = defineProps({
  paginateData: { type: Object as PropType<PaginateData>, required: true },
})

const startPage = computed(() => {
  return Math.floor((props.paginateData.current_page - 1) / 10) * 10 + 1
})

const endPage = computed(() => {
  return Math.min(startPage.value + 9, props.paginateData.last_page)
})

const prevPage = computed(() => {
  return startPage.value > 1 ? startPage.value - 1 : 0
})

const nextPage = computed(() => {
  return endPage.value < props.paginateData.last_page ? endPage.value + 1 : 0
})

const pages = computed(() => {
  return Array.from({ length: endPage.value - startPage.value + 1 }, (_, i) => startPage.value + i)
})

watch(() => props.paginateData, () => { }, { immediate: true })
</script>

<template>
  <div v-if="paginateData.path" class="leading-4 py-2">
    <span v-if="prevPage">
      <RouterLink :to="{ path: `${paginateData.path}/${prevPage}` }" class="btn btn-light !px-2">
        <BaseIcon :path="mdiChevronLeft" />
        이전
      </RouterLink>
    </span>

    <span v-for="page in pages" :key="page">
      <RouterLink :to="{ path: `${paginateData.path}/${page}` }" class="btn btn-light !px-3"
        :class="{ disabled: page == paginateData.current_page }">
        {{ page }}
      </RouterLink>
    </span>

    <span v-if="nextPage">
      <RouterLink :to="{ path: `${paginateData.path}/${nextPage}` }" class="btn btn-light !px-2">
        다음
        <BaseIcon :path="mdiChevronRight" />
      </RouterLink>
    </span>
  </div>
</template>
