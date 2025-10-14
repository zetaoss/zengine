<!-- TheBinder.vue -->
<script setup lang="ts">
import { ref } from 'vue'
import { mdiCog } from '@mdi/js'

import BaseIcon from '@common/ui/BaseIcon.vue'
import CCapSticky from '@common/components/CCapSticky.vue'
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
    const data = await res.json()
    bindersRef.value = data
  } catch (e) {
    console.error(e)
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="binder h-full border-x bg-zinc-50 dark:bg-neutral-900">
    <CCapSticky>
      <div v-if="bindersRef.length">
        <div v-for="binder in bindersRef" :key="binder.id">
          <header class="sticky top-0 z-10 flex items-center justify-between px-3 py-2
                   bg-gray-100 dark:bg-neutral-800 font-bold" @dblclick.stop="refreshBinder">
            <span>{{ binder.title }}</span>
            <a :href="`/wiki/Binder:${binder.title}`"
              class="inline-flex items-center gap-1 rounded-md px-2 py-1 hover:bg-slate-200/70 dark:hover:bg-slate-700/70"
              @click.stop>
              <BaseIcon :path="mdiCog" :class="['h-4 w-4', busy ? 'animate-spin' : '']" />
            </a>
          </header>

          <ul class="m-0 p-0 pt-2 pb-8 list-none">
            <BinderNode v-for="(tree, i) in binder.trees" :key="tree.text" :node="tree" :depth="0"
              :wgArticleId="wgArticleId" :binderId="binder.id" :idx="i" />
          </ul>
        </div>
      </div>
    </CCapSticky>
  </div>
</template>

<style>
.binder {
  ul {
    @apply text-sm tracking-tighter;
  }

  a {
    @apply text-sky-800 dark:text-sky-600;
  }
}
</style>
