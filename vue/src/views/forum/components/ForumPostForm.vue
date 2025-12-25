<!-- ForumPostForm.vue -->
<script setup lang="ts">
import ZButton from '@common/ui/ZButton.vue'
import { computed } from 'vue'

import EditorApex from '../editor/EditorApex.vue'

export interface ForumPostFormValue {
  cat: string
  title: string
  body: string
}

const props = defineProps<{
  modelValue: ForumPostFormValue
  submitText: string
  submitting?: boolean
  titleError?: string | null
  bodyError?: string | null
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: ForumPostFormValue): void
  (e: 'submit'): void
  (e: 'cancel'): void
  (e: 'clearTitleError'): void
  (e: 'clearBodyError'): void
}>()

const disabled = computed(() =>
  !props.modelValue.title.trim() || !props.modelValue.body.trim() || props.submitting
)

const update = (patch: Partial<ForumPostFormValue>) => {
  emit('update:modelValue', { ...props.modelValue, ...patch })
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center">
      <select :value="modelValue.cat" @change="update({ cat: ($event.target as HTMLSelectElement).value })"
        class="my-3 border text-sm rounded w-auto p-1 bg-white dark:bg-black text-gray-900 dark:text-gray-100">
        <option value="질문">질문</option>
        <option value="잡담">잡담</option>
        <option value="인사">인사</option>
        <option value="기타">기타</option>
      </select>
    </div>

    <div>
      <input :value="modelValue.title"
        @input="emit('clearTitleError'); update({ title: ($event.target as HTMLInputElement).value })" type="text"
        placeholder="제목을 입력해 주세요"
        class="border rounded block w-full px-4 py-2 outline-0 bg-white dark:bg-black text-gray-900 dark:text-gray-300"
        :class="{ 'border-red-300 dark:border-red-700': !!titleError }" />
      <div v-if="titleError" class="text-sm text-red-400">
        {{ titleError }}
      </div>
    </div>

    <div>
      <div :class="{ 'border border-red-300 dark:border-red-700': !!bodyError }">
        <EditorApex :model-value="modelValue.body"
          @update:model-value="emit('clearBodyError'); update({ body: $event })" />
      </div>
      <div v-if="bodyError" class="text-sm text-red-400">
        {{ bodyError }}
      </div>
    </div>

    <div class="my-4 flex justify-center gap-3">
      <ZButton color="primary" :disabled="disabled" @click="emit('submit')">
        {{ submitText }}
      </ZButton>
      <ZButton @click="emit('cancel')">
        취소
      </ZButton>
    </div>
  </div>
</template>
