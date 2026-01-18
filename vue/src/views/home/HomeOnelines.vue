<!-- @/views/home/HomeOnelines.vue -->
<script setup lang="ts">
import AvatarUser from '@common/components/avatar/AvatarUser.vue'
import { showConfirm } from '@common/ui/confirm/confirm'
import { showToast } from '@common/ui/toast/toast'
import ZButton from '@common/ui/ZButton.vue'
import ZIcon from '@common/ui/ZIcon.vue'
import httpy from '@common/utils/httpy'
import { mdiDelete } from '@mdi/js'
import { computed, onMounted, ref } from 'vue'

import useAuthStore from '@/stores/auth'
import linkify from '@/utils/linkify'

interface Row {
  id: number
  user_id: number
  user_name: string
  created: string
  message: string
}

const rows = ref<Row[]>([])
const message = ref('')
const isSubmitting = ref(false)
const auth = useAuthStore()
const trimmedMessage = computed(() => message.value.trim())
const canSubmit = computed(() => trimmedMessage.value.length > 0 && !isSubmitting.value)
const canWrite = computed(() => auth.canWrite())

const load = async () => {
  const [data, err] = await httpy.get<Row[]>('/api/onelines/recent')
  if (err) {
    console.error(err)
    return
  }

  rows.value = await Promise.all(
    data.map(async (r) => ({
      ...r,
      message: await linkify(r.message),
    })),
  )
}

const submit = async () => {
  if (!canWrite.value) {
    showToast('작성 불가')
    return
  }
  if (!canSubmit.value) return

  isSubmitting.value = true
  const [data, err] = await httpy.post<Row>('/api/onelines', {
    message: trimmedMessage.value,
  })
  if (err) {
    console.error('create oneline', err)
    showToast(err.message || '등록 실패')
    isSubmitting.value = false
    return
  }

  rows.value = [
    {
      ...data,
      message: await linkify(data.message),
    },
    ...rows.value,
  ]
  message.value = ''
  isSubmitting.value = false
}

const del = async (row: Row) => {
  if (!auth.canDelete(row.user_id)) return
  const ok = await showConfirm('이 한줄잡담을 삭제하시겠습니까?')
  if (!ok) return

  const [, err] = await httpy.delete(`/api/onelines/${row.id}`)
  if (err) {
    console.error(err)
    showToast(err.message || '삭제 실패')
    return
  }

  rows.value = rows.value.filter((item) => item.id !== row.id)
  showToast('삭제 완료')
}

const notifyLogin = () => {
  if (!canWrite.value) {
    showToast('로그인하면 글을 쓸 수 있어요.')
  }
}

onMounted(load)
</script>

<template>
  <form class="mb-2 flex items-center gap-2" @submit.prevent="submit">
    <div class="relative flex-1">
      <input v-model="message" type="text" class="w-full rounded border p-1 px-2" placeholder="What’s on your mind?"
        :disabled="!canWrite || isSubmitting" />
      <button v-if="!canWrite" type="button" aria-label="login required"
        class="absolute inset-0 z-10 cursor-not-allowed" @click="notifyLogin" />
    </div>
    <button type="submit" class="rounded border px-3 py-1 text-sm disabled:opacity-40"
      :disabled="!canWrite || !canSubmit">
      등록
    </button>
  </form>
  <div v-for="r in rows" :key="r.id" class="py-2">
    <AvatarUser :user="{ id: r.user_id, name: r.user_name }" />
    <span class="ml-1" v-html="r.message" />
    <span class="z-muted2 ml-1 text-xs">{{ r.created.substring(0, 10) }}</span>
    <ZButton v-if="auth.canDelete(r.user_id)" color="ghost" class="text-[#888] py-1 align-middle leading-none"
      @click="del(r)">
      <ZIcon :path="mdiDelete" />
    </ZButton>
  </div>
</template>
