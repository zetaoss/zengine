<script setup lang="ts">
import ZIcon from '@common/ui/ZIcon.vue'
import { mdiCog } from '@mdi/js'
import { ref } from 'vue'

import CapSticky from '@/components/CapSticky.vue'
import getRLCONF from '@/utils/rlconf'

import BinderNode from './BinderNode.vue'

const { wgArticleId, binders } = getRLCONF()

const bindersRef = ref(binders ?? [])
const busy = ref(false)

async function refreshBinder() {
  if (busy.value) return
  busy.value = true
  try {
    const res = await fetch(`/w/rest.php/binder/${wgArticleId}?refresh=1`)
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    bindersRef.value = await res.json()
  } catch (e) {
    console.error(e)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <CapSticky :showToggle="true" :widthValue="'240px'" navBlurColor="var(--bg-muted)" class="flex z-bg-muted">
    <div v-for="binder in bindersRef" :key="binder.id">
      <header class="book sticky top-0 z-10 flex items-center justify-between px-3 py-2 bg-gray-200/80 dark:bg-gray-800/80  border-gray-400/60 dark:border-gray-600/60
                 font-bold" @dblclick.stop="refreshBinder">
        <span>{{ binder.title }}</span>
        <a :href="`/wiki/Binder:${binder.title}`" class="inline-flex items-center gap-1 rounded-md px-2 py-1"
          @click.stop>
          <ZIcon :path="mdiCog" :class="['h-4 w-4', busy ? 'animate-spin' : '']" />
        </a>
      </header>

      <ul class="m-0 p-0 pt-2 pb-10 list-none text-[.9rem]">
        <BinderNode v-for="(tree, i) in binder.trees" :key="tree.text" :node="tree" :depth="0"
          :wgArticleId="wgArticleId" :binderId="binder.id" :idx="i" />
      </ul>

    </div>
  </CapSticky>
</template>
