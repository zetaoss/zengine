<script lang="ts">
  import { Editor } from '@tiptap/core'
  import CharacterCount from '@tiptap/extension-character-count'
  import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
  import Highlight from '@tiptap/extension-highlight'
  import Image from '@tiptap/extension-image'
  import Placeholder from '@tiptap/extension-placeholder'
  import { Table, TableCell, TableHeader, TableRow } from '@tiptap/extension-table'
  import TaskItem from '@tiptap/extension-task-item'
  import TaskList from '@tiptap/extension-task-list'
  import StarterKit from '@tiptap/starter-kit'
  import c from 'highlight.js/lib/languages/c'
  import cpp from 'highlight.js/lib/languages/cpp'
  import go from 'highlight.js/lib/languages/go'
  import php from 'highlight.js/lib/languages/php'
  import xml from 'highlight.js/lib/languages/xml'
  import { createLowlight } from 'lowlight'
  import { onDestroy, onMount } from 'svelte'

  import Iframe from './iframe'
  import MenuBar from './MenuBar.svelte'
  import TableMenuBar from './TableMenuBar.svelte'

  export let modelValue = ''
  export let isError = false
  export let onModelValueChange: (value: string) => void = () => {}

  const lowlight = createLowlight()

  lowlight.register('c', c)
  lowlight.register('cpp', cpp)
  lowlight.register('go', go)
  lowlight.register('php', php)
  lowlight.register('html', xml)
  lowlight.register('xml', xml)
  lowlight.register('plaintext', () => ({ contains: [] }))
  lowlight.register('text', () => ({ contains: [] }))

  let editor: Editor | null = null
  let editorEl: HTMLDivElement | null = null
  let hasEditor = false
  let wordCount = 0

  function setContentIfChanged(value: string) {
    if (!editor) return
    const current = editor.getHTML()
    if (current === value) return
    editor.commands.setContent(value, { emitUpdate: false })
  }

  onMount(() => {
    if (!editorEl) return

    editor = new Editor({
      element: editorEl,
      extensions: [
        StarterKit.configure({
          heading: { levels: [1, 2, 3] },
          codeBlock: false,
        }),
        CodeBlockLowlight.configure({
          lowlight,
          defaultLanguage: 'plaintext',
          HTMLAttributes: { class: 'code-block hljs' },
        }),
        Highlight,
        TaskList,
        TaskItem,
        CharacterCount.configure({ limit: 10000 }),
        Table.configure({ resizable: true }),
        TableRow,
        TableHeader,
        TableCell,
        Placeholder.configure({ placeholder: '내용' }),
        Image.configure({ inline: true }),
        Iframe,
      ],
      content: modelValue,
      editorProps: {
        attributes: {
          class: 'tiptap',
        },
      },
      onUpdate: ({ editor: e }) => {
        const html = e.getHTML()
        onModelValueChange(html)
        wordCount = e.storage.characterCount.words()
      },
    })

    hasEditor = true
    wordCount = editor.storage.characterCount.words()
    return () => {
      editor?.destroy()
      editor = null
      hasEditor = false
    }
  })

  onDestroy(() => {
    editor?.destroy()
    editor = null
    hasEditor = false
  })

  $: if (modelValue !== undefined) {
    setContentIfChanged(modelValue)
  }
</script>

<div class={`editor ${isError ? 'border-red-50 dark:border-red-900' : ''}`}>
  {#if hasEditor && editor}
    <MenuBar {editor} />
    <TableMenuBar {editor} />
  {/if}

  <div class="editor__content" bind:this={editorEl}></div>

  {#if hasEditor && editor}
    <div class="editor__footer">
      <div class="w-full text-right">
        <div class="character-count">
          {wordCount} words
        </div>
      </div>
    </div>
  {/if}
</div>

<style lang="postcss">
  @reference 'tailwindcss';

  .editor {
    @apply border rounded bg-white dark:bg-black flex flex-col h-[60vh];
  }

  .editor__content {
    flex: 1 1 auto;
    padding: 1.25rem 1rem;
    -webkit-overflow-scrolling: touch;
    overflow-x: hidden;
    overflow-y: auto;
  }

  .editor__content :global(.tiptap) {
    @apply outline-none;
  }

  .editor__content :global(.tiptap p) {
    @apply leading-7;
  }

  .editor__content :global(.tiptap pre) {
    @apply my-3 rounded bg-zinc-900 text-zinc-100 overflow-x-auto;
    padding: 0.9rem 1rem;
  }

  .editor__content :global(.tiptap pre code) {
    @apply text-sm;
    color: inherit;
  }

  .editor__footer {
    @apply border-t;
    display: flex;
    flex: 0 0 auto;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    padding: 0.25rem 0.75rem;
    font-size: 12px;
    font-weight: 600;
    white-space: nowrap;
  }

  .divider {
    @apply ml-2 mr-1 h-4 w-px bg-gray-300 dark:bg-gray-700;
  }
</style>
