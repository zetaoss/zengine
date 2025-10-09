<script setup lang="ts">
import BaseIcon from '@common/ui/BaseIcon.vue'
import { mdiCog } from '@mdi/js'
import { useWindowScroll } from '@vueuse/core'
import getRLCONF from '@/utils/rlconf'
import BinderNode from './BinderNode.vue'
import { ref } from 'vue'

const { y } = useWindowScroll()
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
  <aside class="p-4 bg-white dark:bg-neutral-900 overflow-y-auto sticky top-0 z-scrollbar border-r tracking-tight"
    :style="{ height: `calc(100vh - ${y > 48 ? 0 : 48 - y}px)` }">
    <div class="rb text-sm" v-if="bindersRef.length">
      <div v-for="binder in bindersRef" :key="binder.id">
        <header class="bg-slate-100 dark:bg-slate-800" @dblclick.stop="refreshBinder">
          {{ binder.title }}
          <a :href="`/wiki/Binder:${binder.title}`" @click.stop>
            <BaseIcon :path="mdiCog" :class="busy ? 'animate-spin' : ''" />
          </a>
        </header>

        <ul class="p-0 m-0" v-for="tree in binder.trees" :key="tree.text">
          <BinderNode :node="tree" />
        </ul>
      </div>
    </div>
  </aside>
</template>
