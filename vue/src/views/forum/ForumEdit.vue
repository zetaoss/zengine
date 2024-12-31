<script setup lang="ts">
import { ref } from 'vue'

import { useRoute, useRouter } from 'vue-router'

import http from '@/utils/http'
import TheModal from '@common/components/TheModal.vue'

import TiptapMain from './tiptap/TiptapMain.vue'
import type { DataError, ErrorResponse } from './types'

import './tiptap/ProseMirror.scss'

const route = useRoute()
const router = useRouter()

const id = ref(0)
const cat = ref('질문')
const title = ref('')
const body = ref('')
const dataError = ref({} as DataError)
const showModal = ref(false)

function gotoPost() {
  router.push({ path: `/forum/${id.value}` })
}

async function fetchData() {
  id.value = Number(route.params.id as string)
  const resp = await http.get(`/api/posts/${id.value}`)
  cat.value = resp.data.cat
  title.value = resp.data.title
  body.value = resp.data.body
}

async function put() {
  if (title.value.length < 1) {
    alert('제목을 입력해 주세요')
    return
  }
  if (body.value.length < 1) {
    alert('내용을 입력해 주세요')
    return
  }
  try {
    await http.put(`/api/posts/${id.value}`, {
      cat: cat.value,
      title: title.value,
      body: body.value,
    })
    gotoPost()
  } catch (err: any) {
    const resp = err?.response as ErrorResponse
    dataError.value = resp.data.error
    console.error(resp.data)
  }
}

function ModalOK() {
  showModal.value = false
  gotoPost()
}

fetchData()
</script>

<template>
  <TheModal :show="showModal" @ok="ModalOK" @cancel="showModal = false">
    글 수정하기를 취소하시겠습니까?
  </TheModal>

  <div class="p-5">
    <div class="container mx-auto px-4 max-w-[1140px]">
      <h2 class="my-5 text-2xl font-bold">
        포럼 글 수정하기
      </h2>
      <div class="flex items-center">
        <select v-model="cat" aria-label="category"
          class="my-3 border text-sm rounded focus:ring-blue-500 focus:border-blue-500 w-auto p-1 bg-white dark:bg-black text-gray-900 dark:text-gray-100">
          <option value="질문">
            질문
          </option>
          <option value="잡담">
            잡담
          </option>
          <option value="인사">
            인사
          </option>
          <option value="기타">
            기타
          </option>
        </select>
        <span class="px-3 py-1 inline-flex items-center">
          <input id="default-checkbox" type="checkbox" value=""
            class="w-5 h-5 rounded focus:ring-blue-500 bg-white dark:bg-black">
          <label for="default-checkbox" class="ml-2 text-sm font-medium text-gray-900 dark:text-gray-300">공지</label>
        </span>
      </div>
      <div>
        <input v-model="title" aria-label="title" type="text" placeholder="제목"
          class="border rounded block w-full px-4 py-2 outline-0 bg-white dark:bg-black text-gray-900 dark:text-gray-300"
          :class="{ 'border-red-300 dark:border-red-700': dataError.title }" @keyup="dataError = {} as DataError">
        <div v-if="dataError.title" class="text-sm text-red-400">
          {{ dataError.title.join('') }}
        </div>
      </div>
      <div class="mt-4">
        <div :class="{ 'border border-red-300 dark:border-red-700': dataError.body }">
          <TiptapMain v-model="body" />
        </div>
        <div v-if="dataError.body" class="rounded text-sm text-red-400">
          {{ dataError.body.join('') }}
        </div>
      </div>
      <div class="mt-4 mb-8 text-center">
        <button type="button" class="btn btn-primary" @click="put">
          등록
        </button>
        <button type="button" class="btn" @click="showModal = true">
          취소
        </button>
      </div>
    </div>
  </div>
</template>
