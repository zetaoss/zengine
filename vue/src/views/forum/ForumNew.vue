<script setup lang="ts">
import './tiptap/ProseMirror.scss'

import ZButton from '@common/ui/ZButton.vue'
import ZModal from '@common/ui/ZModal.vue'
import httpy from '@common/utils/httpy'
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'

import { useErrors } from './errors'
import TiptapMain from './tiptap/TiptapMain.vue'

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

  const [, err] = await httpy.post('/api/posts', {
    cat: cat.value,
    title: title.value,
    body: body.value,
  })
  if (err) {
    console.error('POST posts', err)
    return
  }

  gotoList()
}

const postDisabled = computed(() => {
  return !title.value.trim() || !body.value.trim()
})

const modalOK = () => {
  showModal.value = false
  gotoList()
}
</script>

<template>
  <ZModal :show="showModal" @ok="modalOK" @cancel="showModal = false">
    새 글 쓰기를 취소하시겠습니까?
  </ZModal>

  <div class="p-5">
    <div class="container mx-auto px-4 max-w-[1140px]">
      <h2 class="my-5 text-2xl font-bold">
        포럼 새 글 쓰기
      </h2>

      <div class="flex items-center gap-2">
        <select v-model="cat" class="w-20 border rounded px-2 py-1">
          <option value="질문">질문</option>
          <option value="잡담">잡담</option>
          <option value="인사">인사</option>
          <option value="기타">기타</option>
        </select>
        <input v-model="title" @input="errors.clear('title')" type="text" placeholder="제목을 입력해 주세요"
          class="flex-1 border rounded px-2 py-1"
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
        <ZButton @click="post" color="primary" :disabled="postDisabled">
          등록
        </ZButton>
        <ZButton @click="showModal = true">
          취소
        </ZButton>
      </div>
    </div>
  </div>
</template>
