<!-- eslint-disable vue/multi-word-component-names -->
<script setup lang="ts">
import Icon from '@common/ui/Icon.vue'
import { mdiCog } from '@mdi/js'
import { useWindowScroll } from '@vueuse/core'
import getRLCONF from '@/utils/rlconf'
import BinderNode from './BinderNode.vue'

const { y } = useWindowScroll()
const { binders } = getRLCONF()

function dblclick() {
  console.log('dblclick')
}
</script>
<template>
  <aside class="p-4 bg-white dark:bg-neutral-900 overflow-y-auto sticky top-0 z-scrollbar border-r"
    :style="{ height: `calc(100vh - ${y > 48 ? 0 : 48 - y}px)` }">
    <div class='rb text-sm' v-if='binders.length'>
      <div v-for='(binder) in binders' :key="binder.id">
        <header @dblclick='dblclick'>{{ binder.title }}
          <a :href="`/wiki/Binder:${binder.title}`">
            <Icon :path="mdiCog" />
          </a>
        </header>
        <ul class="p-0 m-0" v-for="tree in binder.trees" :key="tree.text">
          <BinderNode :node='tree' />
        </ul>
      </div>
    </div>
  </aside>
</template>
