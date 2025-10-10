<!-- TheBinder.vue -->
<script setup lang="ts">
import { ref } from 'vue'
import { mdiCog } from '@mdi/js'

import BaseIcon from '@common/ui/BaseIcon.vue'
import CCapSticky from '@common/components/CCapSticky.vue'
import getRLCONF from '@/utils/rlconf'
import BinderNode from './BinderNode.vue'

const { wgArticleId, wgTitle, binders } = getRLCONF()

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
  <div class="h-full border-x bg-zinc-50 dark:bg-neutral-900">
    <CCapSticky>
      <div class="text-xs tracking-tighter" v-if="bindersRef.length">
        <div v-for="binder in bindersRef" :key="binder.id">
          <header class="sticky top-0 z-10 flex items-center justify-between px-3 py-2
                   bg-gray-100 dark:bg-neutral-700 font-bold" @dblclick.stop="refreshBinder">
            <span>{{ binder.title }}</span>
            <a :href="`/wiki/Binder:${binder.title}`"
              class="inline-flex items-center gap-1 rounded-md px-2 py-1 hover:bg-slate-200/70 dark:hover:bg-slate-700/70
                     focus:outline-none focus-visible:ring-2 focus-visible:ring-slate-300 dark:focus-visible:ring-slate-600" @click.stop title="Binder 설정">
              <BaseIcon :path="mdiCog" :class="['h-4 w-4', busy ? 'animate-spin' : '']" />
            </a>
          </header>

          <ul class="m-0 p-1 list-none">
            <BinderNode v-for="tree in binder.trees" :key="tree.text" :node="tree" :depth="0" :wgTitle="wgTitle" />
          </ul>
        </div>
      </div>
    </CCapSticky>
  </div>
</template>
