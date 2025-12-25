<!-- TiptapMain.vue -->
<script setup lang="ts">
import CharacterCount from '@tiptap/extension-character-count'
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import Highlight from '@tiptap/extension-highlight'
import Image from '@tiptap/extension-image'
import Placeholder from '@tiptap/extension-placeholder'
import { Table, TableCell, TableHeader, TableRow } from '@tiptap/extension-table'
import TaskItem from '@tiptap/extension-task-item'
import TaskList from '@tiptap/extension-task-list'
import StarterKit from '@tiptap/starter-kit'
import { Editor, EditorContent } from '@tiptap/vue-3'
import c from 'highlight.js/lib/languages/c'
import cpp from 'highlight.js/lib/languages/cpp'
import go from 'highlight.js/lib/languages/go'
import php from 'highlight.js/lib/languages/php'
import xml from 'highlight.js/lib/languages/xml'
import { createLowlight } from 'lowlight'
import { computed, onBeforeUnmount, onMounted, shallowRef, watch } from 'vue'

import Iframe from './iframe'
import MenuBar from './MenuBar.vue'
import TableMenuBar from './TableMenuBar.vue'

const lowlight = createLowlight()

lowlight.register('c', c)
lowlight.register('cpp', cpp)
lowlight.register('go', go)
lowlight.register('php', php)
lowlight.register('html', xml)
lowlight.register('xml', xml)
lowlight.register('plaintext', () => ({ contains: [] }))
lowlight.register('text', () => ({ contains: [] }))

const props = defineProps({
  modelValue: { type: String, default: '' },
  isError: { type: Boolean, default: false },
})

const emit = defineEmits<{
  (e: 'update:modelValue', v: string): void
}>()

const editor = shallowRef<Editor | null>(null)
const hasEditor = computed(() => !!editor.value)

watch(
  () => props.modelValue,
  (newValue) => {
    const e = editor.value
    if (!e) return

    const current = e.getHTML()
    if (current === newValue) return

    e.commands.setContent(newValue, { emitUpdate: false })

  }
)

onMounted(() => {
  editor.value = new Editor({
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
    content: props.modelValue,
    editorProps: {
      attributes: {
        class: 'tiptap',
      },
    },
    onUpdate: ({ editor: e }) => {
      emit('update:modelValue', e.getHTML())
    },
  })
})

onBeforeUnmount(() => {
  editor.value?.destroy()
  editor.value = null
})
</script>

<template>
  <div v-if="hasEditor" class="editor" :class="{ 'border-red-50 dark:border-red-900': isError }">
    <MenuBar :editor="editor!" />
    <TableMenuBar :editor="editor!" />
    <EditorContent class="editor__content" :editor="editor!" />

    <div class="editor__footer">
      <div class="w-full text-right">
        <div class="character-count">
          {{ editor!.storage.characterCount.words() }} words
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss">
.editor {
  @apply border rounded bg-white dark:bg-black flex flex-col h-[60vh];

  &__content {
    flex: 1 1 auto;
    overflow-x: hidden;
    overflow-y: auto;
    padding: 1.25rem 1rem;
    -webkit-overflow-scrolling: touch;

    .tiptap {
      @apply outline-none;

      p {
        @apply leading-7;
      }

      pre {
        @apply my-3 rounded bg-zinc-900 text-zinc-100 overflow-x-auto;
        padding: 0.9rem 1rem;
      }

      pre code {
        @apply text-sm;
        color: inherit;
      }
    }
  }

  &__footer {
    @apply border-t;
    align-items: center;
    display: flex;
    flex: 0 0 auto;
    flex-wrap: wrap;
    font-size: 12px;
    font-weight: 600;
    justify-content: space-between;
    padding: 0.25rem 0.75rem;
    white-space: nowrap;
  }
}

.divider {
  @apply ml-2 mr-1 h-4 w-[1px] bg-gray-300 dark:bg-gray-700;
}

.menu-item {
  @apply rounded cursor-pointer ml-1;

  svg {
    @apply w-7 h-7 p-1 fill-current;
  }

  &.is-active,
  &:hover {
    @apply bg-zinc-200 dark:bg-zinc-800;
  }
}
</style>
