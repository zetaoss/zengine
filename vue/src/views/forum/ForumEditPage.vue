<!-- ForumEditPage.vue -->
<script setup lang="ts">
import './assets/forum-apex.css'

import ZModal from '@common/ui/ZModal.vue'
import httpy from '@common/utils/httpy'
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import ForumPostForm, { type ForumPostFormValue } from './components/ForumPostForm.vue'
import { useErrors } from './errors'

const route = useRoute()
const router = useRouter()
const errors = useErrors()

const id = computed(() => Number(route.params.id || 0))
const isEdit = computed(() => id.value > 0)

const form = ref<ForumPostFormValue>({
  cat: '질문',
  title: '',
  body: '',
})

const showCancelModal = ref(false)
const submitting = ref(false)
const loading = ref(false)

const titleError = computed(() => (errors.has('title') ? errors.get('title').join('') : null))
const bodyError = computed(() => (errors.has('body') ? errors.get('body').join('') : null))

const fetchData = async () => {
  if (!isEdit.value) return

  loading.value = true
  try {
    const [data, err] = await httpy.get<{ cat: string; title: string; body: string }>(`/api/posts/${id.value}`)
    if (err) {
      console.error('Failed to fetch post:', err)
      return
    }

    form.value = {
      cat: data.cat ?? '질문',
      title: data.title ?? '',
      body: data.body ?? '',
    }
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)

const validate = () => {
  errors.clearAll()

  if (!form.value.title.trim()) errors.add('title', '제목을 입력해 주세요')
  if (!form.value.body.trim()) errors.add('body', '내용을 입력해 주세요')

  return !errors.isError()
}

const gotoAfterSubmit = (newId?: number) => {
  router.push({ path: `/forum/${newId ?? id.value}` })
}

const submit = async () => {
  if (!validate()) return

  submitting.value = true
  try {
    if (isEdit.value) {
      const [, err] = await httpy.put(`/api/posts/${id.value}`, form.value)
      if (err) {
        console.error('PUT post', err)
        return
      }
      gotoAfterSubmit()
      return
    }

    const [data, err] = await httpy.post<{ id: number }>('/api/posts', form.value)
    if (err) {
      console.error('POST post', err)
      return
    }
    gotoAfterSubmit(data?.id)
  } finally {
    submitting.value = false
  }
}

const cancel = () => {
  showCancelModal.value = true
}

const cancelOk = () => {
  showCancelModal.value = false
  router.push({ path: isEdit.value ? `/forum/${id.value}` : '/forum' })
}
</script>

<template>
  <ZModal :show="showCancelModal" @ok="cancelOk" @cancel="showCancelModal = false">
    {{ isEdit ? '글 수정하기를 취소하시겠습니까?' : '새 글 쓰기를 취소하시겠습니까?' }}
  </ZModal>

  <div class="p-5">
    <div class="container mx-auto px-4 max-w-[1140px]">
      <h2 class="my-5 text-2xl font-bold">
        {{ isEdit ? '포럼 글 수정하기' : '포럼 새 글 쓰기' }}
      </h2>

      <div v-if="loading" class="py-10 text-center text-gray-500">
        불러오는 중...
      </div>

      <ForumPostForm v-else v-model="form" submit-text="저장" :submitting="submitting" :title-error="titleError"
        :body-error="bodyError" @submit="submit" @cancel="cancel" @clear-title-error="errors.clear('title')"
        @clear-body-error="errors.clear('body')" />
    </div>
  </div>
</template>
