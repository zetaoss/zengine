<script setup lang="ts">
import { onBeforeUnmount, onMounted, shallowRef, watch } from 'vue'

import CharacterCount from '@tiptap/extension-character-count'
import Highlight from '@tiptap/extension-highlight'
import Image from '@tiptap/extension-image'
import Placeholder from '@tiptap/extension-placeholder'
import { Table, TableRow, TableCell, TableHeader } from '@tiptap/extension-table'
import TaskItem from '@tiptap/extension-task-item'
import TaskList from '@tiptap/extension-task-list'
import StarterKit from '@tiptap/starter-kit'
import { Editor, EditorContent } from '@tiptap/vue-3'

import MenuBar from './MenuBar.vue'
import TableMenuBar from './TableMenuBar.vue'
import Iframe from './iframe'

const props = defineProps({
  modelValue: { type: String, default: '' },
  isError: { type: Boolean, default: false },
})

const emit = defineEmits(['update:modelValue'])
const editor = shallowRef<Editor | null>(null)

watch(() => props.modelValue, (newValue) => {
  if (editor.value && editor.value.getHTML() !== newValue) {
    editor.value.commands.setContent(newValue)
  }
})

onMounted(() => {
  editor.value = new Editor({
    extensions: [
      StarterKit.configure({
        heading: { levels: [1, 2, 3] },
      }),
      Highlight,
      TaskList,
      TaskItem,
      CharacterCount.configure({ limit: 10000 }),
      Table.configure({ resizable: true }),
      TableCell,
      TableHeader,
      TableRow,
      Placeholder.configure({ placeholder: '내용' }),
      Image.configure({ inline: true }),
      Iframe,
    ],
    content: props.modelValue,
    onUpdate: () => {
      if (editor.value) {
        emit('update:modelValue', editor.value.getHTML())
      }
    },
  })
})

onBeforeUnmount(() => {
  editor.value?.destroy()
})
</script>

<template>
  <div v-if="editor" class="editor" :class="{ 'border-red-50 dark:border-red-900': isError }">
    <menu-bar :editor="editor" />
    <table-menu-bar :editor="editor" />
    <editor-content v-if="editor" class="editor__content" :editor="editor" />
    <div class="editor__footer">
      <div class="w-full text-right">
        <div v-if="editor" class="character-count">
          {{ editor.storage.characterCount.words() }} words
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

// menu bar
.divider {
  @apply ml-2 mr-1 h-4 w-[1px] bg-gray-300 dark:bg-gray-700;
}

// menu item
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
