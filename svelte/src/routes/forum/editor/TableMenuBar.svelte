<script lang="ts">
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
  import type { Editor } from '@tiptap/core'
  import { onDestroy } from 'svelte'

  import MenuItem from './MenuItem.svelte'
  import type { ItemData } from './types'

  export let editor: Editor

  const itemDatas: ItemData[] = [
    {
      icon: mdiTableRemove,
      title: 'Delete Table',
      action: () => editor.chain().focus().deleteTable().run(),
    },
    { type: 'divider' },
    {
      icon: mdiTableRowPlusBefore,
      title: 'Add Row Above',
      action: () => editor.chain().focus().addRowBefore().run(),
    },
    {
      icon: mdiTableRowPlusAfter,
      title: 'Add Row Below',
      action: () => editor.chain().focus().addRowAfter().run(),
    },
    {
      icon: mdiTableRowRemove,
      title: 'Delete Row',
      action: () => editor.chain().focus().deleteRow().run(),
    },
    { type: 'divider' },
    {
      icon: mdiTableColumnPlusBefore,
      title: 'Add Column Left',
      action: () => editor.chain().focus().addColumnBefore().run(),
    },
    {
      icon: mdiTableColumnPlusAfter,
      title: 'Add Column Right',
      action: () => editor.chain().focus().addColumnAfter().run(),
    },
    {
      icon: mdiTableColumnRemove,
      title: 'Delete Column',
      action: () => editor.chain().focus().deleteColumn().run(),
    },
    { type: 'divider' },
    {
      icon: mdiTableMergeCells,
      title: 'Merge/Split Cells',
      action: () => editor.chain().focus().mergeOrSplit().run(),
    },
    { type: 'divider' },
    {
      icon: mdiDockTop,
      title: 'Toggle Header Row',
      action: () => editor.chain().focus().toggleHeaderRow().run(),
    },
    {
      icon: mdiDockLeft,
      title: 'Toggle Header Column',
      action: () => editor.chain().focus().toggleHeaderColumn().run(),
    },
    {
      icon: mdiCheckboxBlank,
      title: 'Toggle Header Cell',
      action: () => editor.chain().focus().toggleHeaderCell().run(),
    },
  ]

  let show = false

  function updateShow() {
    show = editor.isActive('table') || editor.isActive('tableCell') || editor.isActive('tableHeader')
  }

  if (editor) {
    editor.on('selectionUpdate', updateShow)
    editor.on('transaction', updateShow)
    updateShow()
  }

  onDestroy(() => {
    if (!editor) return
    editor.off('selectionUpdate', updateShow)
    editor.off('transaction', updateShow)
  })
</script>

{#if show}
  <div class="table-menu-bar">
    {#each itemDatas as itemData, index (index)}
      {#if itemData.type === 'divider'}
        <div class="divider"></div>
      {:else}
        <MenuItem {itemData} />
      {/if}
    {/each}
  </div>
{/if}

<style>
  @reference 'tailwindcss';

  .table-menu-bar {
    @apply border-b text-right items-center flex flex-wrap p-1 bg-slate-200 dark:bg-slate-800;
  }

  .divider {
    @apply ml-2 mr-1 h-4 w-px bg-gray-300 dark:bg-gray-700;
  }
</style>
