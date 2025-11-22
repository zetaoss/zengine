<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import http from '@/utils/http'
import BaseModal from '@common/ui/BaseModal.vue'
import TiptapMain from './tiptap/TiptapMain.vue'
import { useErrors, type ErrorResponse } from './errors'
import './tiptap/ProseMirror.scss'
import UiButton from '@common/ui/UiButton.vue'

const router = useRouter()
const errors = useErrors()

const cat = ref<string>('질문')
const title = ref<string>('')
const body = ref<string>('')
const showModal = ref(false)

const gotoList = () => {
  router.push({ path: '/forum' })
}

const post = async () => {
  errors.clearAll()

  if (!title.value.trim()) {
    errors.add('title', '제목을 입력해 주세요')
  }
  if (!body.value.trim()) {
    errors.add('body', '내용을 입력해 주세요')
  }
  if (errors.isError()) return

  try {
    await http.post('/api/posts', {
      cat: cat.value,
      title: title.value,
      body: body.value,
    })
    gotoList()
  } catch (err) {
    if (err instanceof Error && 'response' in err) {
      const resp = err?.response as ErrorResponse
      Object.entries(resp?.data?.error || {}).forEach(([field, messages]) => {
        messages.forEach((msg) => errors.add(field, msg))
      })
    } else {
      console.error('API request error:', err)
    }
  }
}

const modalOK = () => {
  showModal.value = false
  gotoList()
}
</script>

<template>
  <BaseModal :show="showModal" @ok="modalOK" @cancel="showModal = false">
    새 글 쓰기를 취소하시겠습니까?
  </BaseModal>

  <div class="p-5">
    <div class="container mx-auto px-4 max-w-[1140px]">
      <h2 class="my-5 text-2xl font-bold">
        포럼 새 글 쓰기
      </h2>

      <div class="flex items-center">
        <select v-model="cat" aria-label="category"
          class="my-3 border text-sm rounded focus:ring-blue-500 focus:border-blue-500 w-auto p-1 bg-white dark:bg-black text-gray-900 dark:text-gray-100">
          <option value="질문">질문</option>
          <option value="잡담">잡담</option>
          <option value="인사">인사</option>
          <option value="기타">기타</option>
        </select>
      </div>

      <div>
        <input v-model="title" @input="errors.clear('title')" type="text" placeholder="제목을 입력해 주세요"
          class="border rounded block w-full px-4 py-2 outline-0 bg-white dark:bg-black text-gray-900 dark:text-gray-300"
          :class="{ 'border-red-300 dark:border-red-700': errors.has('title') }">
        <div v-if="errors.has('title')" class="text-sm text-red-400">
          {{ errors.get('title').join('') }}
        </div>
      </div>

      <div class="mt-4">
        <div :class="{ 'border border-red-300 dark:border-red-700': errors.has('body') }">
          <TiptapMain v-model="body" @update:model-value="errors.clear('body')" />
        </div>
        <div v-if="errors.has('body')" class="text-sm text-red-400">
          {{ errors.get('body').join('') }}
        </div>
      </div>

      <div class="my-4 flex justify-center space-x-3">
        <UiButton @click="post">등록</UiButton>
        <UiButton @click="showModal = true">취소</UiButton>
      </div>
    </div>
  </div>
</template>
