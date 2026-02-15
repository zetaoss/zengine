<script lang="ts">
  import {
    mdiArrowULeftTop,
    mdiArrowURightTop,
    mdiCodeBraces,
    mdiCodeBracesBox,
    mdiFormatBold,
    mdiFormatClear,
    mdiFormatColorHighlight,
    mdiFormatHeader1,
    mdiFormatHeader2,
    mdiFormatHeader3,
    mdiFormatItalic,
    mdiFormatListBulleted,
    mdiFormatListNumbered,
    mdiFormatListText,
    mdiFormatParagraph,
    mdiFormatQuoteOpen,
    mdiFormatStrikethroughVariant,
    mdiImageOutline,
    mdiMinus,
    mdiPlayBoxOutline,
    mdiTable,
    mdiWrap,
  } from '@mdi/js'
  import type { Editor } from '@tiptap/core'

  import MenuItem from './MenuItem.svelte'
  import type { ItemData } from './types'

  export let editor: Editor

  type CodeLang = 'plaintext' | 'go' | 'php' | 'html' | 'xml' | 'c' | 'cpp'

  const CODE_LANGS: { value: CodeLang; label: string }[] = [
    { value: 'plaintext', label: 'Plain' },
    { value: 'go', label: 'Go' },
    { value: 'php', label: 'PHP' },
    { value: 'html', label: 'HTML' },
    { value: 'xml', label: 'XML' },
    { value: 'c', label: 'C' },
    { value: 'cpp', label: 'C++' },
  ]

  $: currentCodeLang = (() => {
    const lang = (editor.getAttributes('codeBlock')?.language ?? 'plaintext') as string
    const ok = CODE_LANGS.some((x) => x.value === lang)
    return (ok ? lang : 'plaintext') as CodeLang
  })()

  function setCodeBlockLanguage(lang: CodeLang) {
    editor.chain().focus().updateAttributes('codeBlock', { language: lang }).run()
  }

  function toggleCodeBlockWithLang(lang: CodeLang) {
    if (editor.isActive('codeBlock')) {
      editor.chain().focus().toggleCodeBlock().run()
      return
    }

    editor.chain().focus().toggleCodeBlock({ language: lang }).run()
  }

  const itemDatas: ItemData[] = [
    {
      icon: mdiArrowULeftTop,
      title: 'Undo',
      action: () => editor.chain().focus().undo().run(),
    },
    {
      icon: mdiArrowURightTop,
      title: 'Redo',
      action: () => editor.chain().focus().redo().run(),
    },
    { type: 'divider' },

    {
      icon: mdiFormatHeader1,
      title: 'Heading 1',
      action: () => editor.chain().focus().toggleHeading({ level: 1 }).run(),
      isActive: () => editor.isActive('heading', { level: 1 }),
    },
    {
      icon: mdiFormatHeader2,
      title: 'Heading 2',
      action: () => editor.chain().focus().toggleHeading({ level: 2 }).run(),
      isActive: () => editor.isActive('heading', { level: 2 }),
    },
    {
      icon: mdiFormatHeader3,
      title: 'Heading 3',
      action: () => editor.chain().focus().toggleHeading({ level: 3 }).run(),
      isActive: () => editor.isActive('heading', { level: 3 }),
    },
    {
      icon: mdiFormatParagraph,
      title: 'Paragraph',
      action: () => editor.chain().focus().setParagraph().run(),
      isActive: () => editor.isActive('paragraph'),
    },
    { type: 'divider' },

    {
      icon: mdiFormatBold,
      title: 'Bold',
      action: () => editor.chain().focus().toggleBold().run(),
      isActive: () => editor.isActive('bold'),
    },
    {
      icon: mdiFormatItalic,
      title: 'Italic',
      action: () => editor.chain().focus().toggleItalic().run(),
      isActive: () => editor.isActive('italic'),
    },
    {
      icon: mdiFormatStrikethroughVariant,
      title: 'Strike',
      action: () => editor.chain().focus().toggleStrike().run(),
      isActive: () => editor.isActive('strike'),
    },
    {
      icon: mdiFormatColorHighlight,
      title: 'Highlight',
      action: () => editor.chain().focus().toggleHighlight().run(),
      isActive: () => editor.isActive('highlight'),
    },
    {
      icon: mdiCodeBraces,
      title: 'Code',
      action: () => editor.chain().focus().toggleCode().run(),
      isActive: () => editor.isActive('code'),
    },
    {
      icon: mdiFormatClear,
      title: 'Clear Format',
      action: () => editor.chain().focus().clearNodes().unsetAllMarks().run(),
    },
    { type: 'divider' },

    {
      icon: mdiFormatListBulleted,
      title: 'Bullet List',
      action: () => editor.chain().focus().toggleBulletList().run(),
      isActive: () => editor.isActive('bulletList'),
    },
    {
      icon: mdiFormatListNumbered,
      title: 'Ordered List',
      action: () => editor.chain().focus().toggleOrderedList().run(),
      isActive: () => editor.isActive('orderedList'),
    },
    {
      icon: mdiFormatListText,
      title: 'Task List',
      action: () => editor.chain().focus().toggleTaskList().run(),
      isActive: () => editor.isActive('taskList'),
    },

    {
      icon: mdiCodeBracesBox,
      title: 'Code Block',
      action: () => toggleCodeBlockWithLang(currentCodeLang),
      isActive: () => editor.isActive('codeBlock'),
    },
    {
      icon: mdiFormatQuoteOpen,
      title: 'Blockquote',
      action: () => editor.chain().focus().toggleBlockquote().run(),
      isActive: () => editor.isActive('blockquote'),
    },
    { type: 'divider' },

    {
      icon: mdiImageOutline,
      title: 'Image',
      action: () => {
        const url = window.prompt('Image URL')
        if (url) editor.chain().focus().setImage({ src: url }).run()
      },
    },
    {
      icon: mdiPlayBoxOutline,
      title: 'Media',
      action: () => {
        const url = window.prompt('Media URL')
        if (url) editor.chain().focus().setIframe({ src: url }).run()
      },
    },
    {
      icon: mdiTable,
      title: 'Table',
      action: () => editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run(),
      isActive: () => editor.isActive('table'),
    },
    { type: 'divider' },

    {
      icon: mdiWrap,
      title: 'Hard Break',
      action: () => editor.chain().focus().setHardBreak().run(),
    },
    {
      icon: mdiMinus,
      title: 'Horizontal Rule',
      action: () => editor.chain().focus().setHorizontalRule().run(),
    },
  ]
</script>

<div class="menu-bar">
  {#each itemDatas as itemData, index (index)}
    {#if itemData.type === 'divider'}
      <div class="divider"></div>
    {:else}
      <MenuItem {itemData} />
    {/if}
  {/each}

  <div class="divider"></div>

  <div class="code-lang">
    <select
      class="code-lang__select"
      value={currentCodeLang}
      disabled={!editor.isActive('codeBlock')}
      on:change={(e) => setCodeBlockLanguage((e.target as HTMLSelectElement).value as CodeLang)}
    >
      {#each CODE_LANGS as l (l.value)}
        <option value={l.value}>{l.label}</option>
      {/each}
    </select>
  </div>
</div>

<style>
  @reference 'tailwindcss';

  .menu-bar {
    @apply border-b text-right items-center flex flex-wrap p-1 bg-slate-100 dark:bg-slate-900;
  }

  .code-lang {
    @apply ml-auto flex items-center;
  }

  .code-lang__select {
    @apply text-xs rounded border px-2 py-1 bg-white dark:bg-slate-800 dark:text-slate-100 dark:border-slate-700;
  }

  .divider {
    @apply ml-2 mr-1 h-4 w-px bg-gray-300 dark:bg-gray-700;
  }
</style>
