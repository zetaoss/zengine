<!-- TableMenuBar.vue -->
<script setup lang="ts">
import {
  mdiCheckboxBlank,
  mdiDockLeft,
  mdiDockTop,
  mdiTableColumnPlusAfter,
  mdiTableColumnPlusBefore,
  mdiTableColumnRemove,
  mdiTableMergeCells,
  mdiTableRemove,
  mdiTableRowPlusAfter,
  mdiTableRowPlusBefore,
  mdiTableRowRemove,
} from '@mdi/js'
import { onBeforeUnmount, onMounted, ref } from 'vue'

import MenuItem from './MenuItem.vue'
import type { ItemData } from './types'

const props = defineProps({
  editor: { type: Object, required: true },
})

const itemDatas = ref<ItemData[]>([
  {
    icon: mdiTableRemove,
    title: 'Delete Table',
    action: () => props.editor.chain().focus().deleteTable().run(),
  },
  { type: 'divider' },
  {
    icon: mdiTableRowPlusBefore,
    title: 'Add Row Above',
    action: () => props.editor.chain().focus().addRowBefore().run(),
  },
  {
    icon: mdiTableRowPlusAfter,
    title: 'Add Row Below',
    action: () => props.editor.chain().focus().addRowAfter().run(),
  },
  {
    icon: mdiTableRowRemove,
    title: 'Delete Row',
    action: () => props.editor.chain().focus().deleteRow().run(),
  },
  { type: 'divider' },
  {
    icon: mdiTableColumnPlusBefore,
    title: 'Add Column Left',
    action: () => props.editor.chain().focus().addColumnBefore().run(),
  },
  {
    icon: mdiTableColumnPlusAfter,
    title: 'Add Column Right',
    action: () => props.editor.chain().focus().addColumnAfter().run(),
  },
  {
    icon: mdiTableColumnRemove,
    title: 'Delete Column',
    action: () => props.editor.chain().focus().deleteColumn().run(),
  },
  { type: 'divider' },
  {
    icon: mdiTableMergeCells,
    title: 'Merge/Split Cells',
    action: () => props.editor.chain().focus().mergeOrSplit().run(),
  },
  { type: 'divider' },
  {
    icon: mdiDockTop,
    title: 'Toggle Header Row',
    action: () => props.editor.chain().focus().toggleHeaderRow().run(),
  },
  {
    icon: mdiDockLeft,
    title: 'Toggle Header Column',
    action: () => props.editor.chain().focus().toggleHeaderColumn().run(),
  },
  {
    icon: mdiCheckboxBlank,
    title: 'Toggle Header Cell',
    action: () => props.editor.chain().focus().toggleHeaderCell().run(),
  },
])

const show = ref(false)

const updateShow = () => {
  const e = props.editor
  show.value = e.isActive('table') || e.isActive('tableCell') || e.isActive('tableHeader')
}

onMounted(() => {
  const e = props.editor
  e.on('selectionUpdate', updateShow)
  e.on('transaction', updateShow)

  updateShow()
})

onBeforeUnmount(() => {
  const e = props.editor
  e.off('selectionUpdate', updateShow)
  e.off('transaction', updateShow)
})
</script>

<template>
  <div v-if="show" class="table-menu-bar">
    <template v-for="(itemData, index) in itemDatas" :key="index">
      <div v-if="itemData.type === 'divider'" class="divider" />
      <MenuItem v-else :item-data="itemData" />
    </template>
  </div>
</template>

<style scoped>
.table-menu-bar {
  @apply border-b text-right items-center flex flex-wrap p-1 bg-slate-200 dark:bg-slate-800;
}

.menu-item {

  &.is-active,
  &:hover {
    @apply bg-zinc-300 dark:bg-zinc-600;
  }
}
</style>
