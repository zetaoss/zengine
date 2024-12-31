<script setup lang="ts">
import { ref, watch } from 'vue'

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

import MenuItem from './MenuItem.vue'
import type { ItemData } from './types'

const props = defineProps({
  editor: { type: Object, required: true },
})

const itemDatas = ref([
  {
    icon: mdiTableRemove,
    title: 'Delete Table',
    action: () => props.editor.chain().focus().deleteTable().run(),
  },
  {
    type: 'divider',
  },
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
  {
    type: 'divider',
  },
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
  {
    type: 'divider',
  },
  {
    icon: mdiTableMergeCells,
    title: 'Toggle Header Cell',
    action: () => props.editor.chain().focus().mergeOrSplit().run(),
  },
  {
    type: 'divider',
  },
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
] as ItemData[])

const show = ref(false)

let firstTime = true

// TODO: check selectionUpdate table
watch(() => props.editor, (value, newValue) => {
  if (firstTime || value !== newValue) {
    firstTime = false
    props.editor.on('selectionUpdate', () => {
      show.value = props.editor.isActive('table')
    })
  }
})
</script>

<template>
  <div v-if="show" class="table-menu-bar">
    <template v-for="(itemData, index) in itemDatas">
      <div v-if="itemData.type === 'divider'" :key="`divider${index}`" class="divider" />
      <menu-item v-else :key="index" :item-data="itemData" />
    </template>
  </div>
</template>

<style scoped lang="scss">
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
