<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import http from '@/utils/http'
import TheModal from '@common/ui/Modal.vue'
import TiptapMain from './tiptap/TiptapMain.vue'
import { useErrors, type ErrorResponse } from './errors'
import './tiptap/ProseMirror.scss'

const route = useRoute()
const router = useRouter()
const errors = useErrors()

const id = ref<number>(0)
const cat = ref<string>('질문')
const title = ref<string>('')
const body = ref<string>('')
const showModal = ref<boolean>(false)

const gotoPost = () => {
  router.push({ path: `/forum/${id.value}` })
}

const fetchData = async () => {
  try {
    id.value = Number(route.params.id)
    const { data } = await http.get(`/api/posts/${id.value}`)
    cat.value = data.cat
    title.value = data.title
    body.value = data.body
  } catch (error) {
    console.error('Failed to fetch data:', error)
  }
}


const put = async () => {
  errors.clearAll()

  if (!title.value.trim()) {
    errors.add('title', '제목을 입력해 주세요')
  }
  if (!body.value.trim()) {
    errors.add('body', '내용을 입력해 주세요')
  }
  if (errors.isError()) return

  try {
    await http.put(`/api/posts/${id.value}`, {
      cat: cat.value,
      title: title.value,
      body: body.value
    })
    gotoPost()
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

const ModalOK = () => {
  showModal.value = false
  gotoPost()
}

onMounted(fetchData)
</script>

<template>
  <TheModal :show="showModal" @ok="ModalOK" @cancel="showModal = false">
    글 수정하기를 취소하시겠습니까?
  </TheModal>

  <div class="p-5">
    <div class="container mx-auto px-4 max-w-[1140px]">
      <h2 class="my-5 text-2xl font-bold">포럼 글 수정하기</h2>

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

      <div class="mt-4 mb-8 text-center">
        <button type="button" class="btn btn-primary" @click="put">등록</button>
        <button type="button" class="btn" @click="showModal = true">취소</button>
      </div>
    </div>
  </div>
</template>
