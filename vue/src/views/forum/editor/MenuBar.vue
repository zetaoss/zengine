<!-- MenuBar.vue -->
<script setup lang="ts">
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
import type { Editor } from '@tiptap/vue-3'
import { computed, ref } from 'vue'

import MenuItem from './MenuItem.vue'
import type { ItemData } from './types'

type CodeLang = 'plaintext' | 'go' | 'php' | 'html' | 'xml' | 'c' | 'cpp'

const props = defineProps<{
  editor: Editor
}>()

const CODE_LANGS: { value: CodeLang; label: string }[] = [
  { value: 'plaintext', label: 'Plain' },
  { value: 'go', label: 'Go' },
  { value: 'php', label: 'PHP' },
  { value: 'html', label: 'HTML' },
  { value: 'xml', label: 'XML' },
  { value: 'c', label: 'C' },
  { value: 'cpp', label: 'C++' },
]

const currentCodeLang = computed<CodeLang>(() => {
  const e = props.editor
  const lang = (e.getAttributes('codeBlock')?.language ?? 'plaintext') as string
  const ok = CODE_LANGS.some((x) => x.value === lang)
  return (ok ? lang : 'plaintext') as CodeLang
})

function setCodeBlockLanguage(lang: CodeLang) {
  const e = props.editor
  if (!e) return

  e.chain().focus().updateAttributes('codeBlock', { language: lang }).run()
}

function toggleCodeBlockWithLang(lang: CodeLang) {
  const e = props.editor
  if (!e) return

  if (e.isActive('codeBlock')) {
    e.chain().focus().toggleCodeBlock().run()
    return
  }

  e.chain().focus().toggleCodeBlock({ language: lang }).run()
}

const itemDatas = ref([
  {
    icon: mdiArrowULeftTop,
    title: 'Undo',
    action: () => props.editor.chain().focus().undo().run(),
  },
  {
    icon: mdiArrowURightTop,
    title: 'Redo',
    action: () => props.editor.chain().focus().redo().run(),
  },
  { type: 'divider' },

  {
    icon: mdiFormatHeader1,
    title: 'Heading 1',
    action: () => props.editor.chain().focus().toggleHeading({ level: 1 }).run(),
    isActive: () => props.editor.isActive('heading', { level: 1 }),
  },
  {
    icon: mdiFormatHeader2,
    title: 'Heading 2',
    action: () => props.editor.chain().focus().toggleHeading({ level: 2 }).run(),
    isActive: () => props.editor.isActive('heading', { level: 2 }),
  },
  {
    icon: mdiFormatHeader3,
    title: 'Heading 3',
    action: () => props.editor.chain().focus().toggleHeading({ level: 3 }).run(),
    isActive: () => props.editor.isActive('heading', { level: 3 }),
  },
  {
    icon: mdiFormatParagraph,
    title: 'Paragraph',
    action: () => props.editor.chain().focus().setParagraph().run(),
    isActive: () => props.editor.isActive('paragraph'),
  },
  { type: 'divider' },

  {
    icon: mdiFormatBold,
    title: 'Bold',
    action: () => props.editor.chain().focus().toggleBold().run(),
    isActive: () => props.editor.isActive('bold'),
  },
  {
    icon: mdiFormatItalic,
    title: 'Italic',
    action: () => props.editor.chain().focus().toggleItalic().run(),
    isActive: () => props.editor.isActive('italic'),
  },
  {
    icon: mdiFormatStrikethroughVariant,
    title: 'Strike',
    action: () => props.editor.chain().focus().toggleStrike().run(),
    isActive: () => props.editor.isActive('strike'),
  },
  {
    icon: mdiFormatColorHighlight,
    title: 'Highlight',
    action: () => props.editor.chain().focus().toggleHighlight().run(),
    isActive: () => props.editor.isActive('highlight'),
  },
  {
    icon: mdiCodeBraces,
    title: 'Code',
    action: () => props.editor.chain().focus().toggleCode().run(),
    isActive: () => props.editor.isActive('code'),
  },
  {
    icon: mdiFormatClear,
    title: 'Clear Format',
    action: () =>
      props.editor
        .chain()
        .focus()
        .clearNodes()
        .unsetAllMarks()
        .run(),
  },
  { type: 'divider' },

  {
    icon: mdiFormatListBulleted,
    title: 'Bullet List',
    action: () => props.editor.chain().focus().toggleBulletList().run(),
    isActive: () => props.editor.isActive('bulletList'),
  },
  {
    icon: mdiFormatListNumbered,
    title: 'Ordered List',
    action: () => props.editor.chain().focus().toggleOrderedList().run(),
    isActive: () => props.editor.isActive('orderedList'),
  },
  {
    icon: mdiFormatListText,
    title: 'Task List',
    action: () => props.editor.chain().focus().toggleTaskList().run(),
    isActive: () => props.editor.isActive('taskList'),
  },

  {
    icon: mdiCodeBracesBox,
    title: 'Code Block',
    action: () => toggleCodeBlockWithLang(currentCodeLang.value),
    isActive: () => props.editor.isActive('codeBlock'),
  },
  {
    icon: mdiFormatQuoteOpen,
    title: 'Blockquote',
    action: () => props.editor.chain().focus().toggleBlockquote().run(),
    isActive: () => props.editor.isActive('blockquote'),
  },
  { type: 'divider' },

  {
    icon: mdiImageOutline,
    title: 'Image',
    action: () => {
      const url = window.prompt('Image URL')
      if (url) props.editor.chain().focus().setImage({ src: url }).run()
    },
  },
  {
    icon: mdiPlayBoxOutline,
    title: 'Media',
    action: () => {
      const url = window.prompt('Media URL')
      if (url) props.editor.chain().focus().setIframe({ src: url }).run()
    },
  },
  {
    icon: mdiTable,
    title: 'Table',
    action: () => props.editor.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run(),
    isActive: () => props.editor.isActive('table'),
  },
  { type: 'divider' },

  {
    icon: mdiWrap,
    title: 'Hard Break',
    action: () => props.editor.chain().focus().setHardBreak().run(),
  },
  {
    icon: mdiMinus,
    title: 'Horizontal Rule',
    action: () => props.editor.chain().focus().setHorizontalRule().run(),
  },
] as ItemData[])
</script>

<template>
  <div class="menu-bar">
    <template v-for="(itemData, index) in itemDatas" :key="index">
      <template v-if="itemData.type === 'divider'">
        <div class="divider" />
      </template>
      <template v-else>
        <MenuItem :item-data="itemData" />
      </template>
    </template>

    <div class="divider" />

    <div class="code-lang">
      <select class="code-lang__select" :value="currentCodeLang" :disabled="!editor.isActive('codeBlock')"
        @change="setCodeBlockLanguage(($event.target as HTMLSelectElement).value as any)">
        <option v-for="l in CODE_LANGS" :key="l.value" :value="l.value">
          {{ l.label }}
        </option>
      </select>
    </div>
  </div>
</template>

<style scoped>
.menu-bar {
  @apply border-b text-right items-center flex flex-wrap p-1 bg-slate-100 dark:bg-slate-900;
}

.code-lang {
  @apply ml-auto flex items-center;
}

.code-lang__select {
  @apply text-xs rounded border px-2 py-1 bg-white dark:bg-slate-800 dark:text-slate-100 dark:border-slate-700;
}
</style>
