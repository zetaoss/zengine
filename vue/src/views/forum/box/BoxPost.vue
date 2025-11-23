<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'

import { useDateFormat } from '@vueuse/core'
import Prism from 'prismjs'
import { useRouter } from 'vue-router'

import ZModal from '@common/ui/ZModal.vue'
import ZButton from '@common/ui/ZButton.vue'
import AvatarUser from '@common/components/avatar/AvatarUser.vue'
import useAuthStore from '@/stores/auth'
import http from '@/utils/http'

import type { Post } from '../types'

import BoxHTML from './BoxHTML.vue'
import BoxReplies from './BoxReplies.vue'

import 'prismjs/components/prism-c'
import 'prismjs/components/prism-cpp'
import 'prismjs/components/prism-go'
import 'prismjs/components/prism-markup-templating'
import 'prismjs/components/prism-php'
import 'prismjs/themes/prism-okaidia.css'

const props = defineProps({
  postID: { type: Number, default: 0 },
})

const router = useRouter()
const auth = useAuthStore()

const post = ref({} as Post)
const showModal = ref(false)

async function fetchData() {
  const resp = await http.get(`/api/posts/${props.postID}`)
  post.value = resp.data
}

function edit() {
  router.push({ path: `/forum/${props.postID}/edit` })
}

function del() {
  showModal.value = true
}

async function modalOK() {
  showModal.value = false
  await http.delete(`/api/posts/${props.postID}`)
  window.location.href = '/forum'
}

watch(() => props.postID, async () => {
  await fetchData()
  await nextTick()
  Prism.highlightAll()
})

fetchData()
</script>

<template>
  <ZModal :show="showModal" @ok="modalOK" @cancel="showModal = false">
    글을 삭제하시겠습니까?
  </ZModal>
  <div v-if="post">
    <div class="border rounded py-4 z-card">
      <div class="px-4">
        <div>
          <div class="float-left mr-2">
            <span class="text-xs p-1 rounded border text-gray-600 dark:text-gray-400">{{ post.cat }}</span>
          </div>
          <h3 class="text-xl">
            {{ post.title }}
          </h3>
        </div>
        <div v-if="post.userAvatar" class="py-3">
          <div>
            <AvatarUser :user-avatar="post.userAvatar" />
          </div>
          <div class="text-sm">
            <span class="mr-4">{{ useDateFormat(post.created_at, 'YYYY-MM-DD HH:mm').value }}</span>
            <span>조회 {{ post.hit }}</span>
          </div>
        </div>
        <hr>
        <div v-if="post.body">
          <BoxHTML class="py-4 min-h-[9rem]" :body="post.body" />
        </div>
        <hr>
      </div>
      <div>
        <BoxReplies :post-i-d="postID" />
      </div>
    </div>
    <div>
      <div class="py-4 flex gap-3">
        <ZButton :to="{ path: `/forum/new` }" :disabled="!auth.canWrite()">글쓰기</ZButton>
        <template v-if="auth.canEdit(post.user_id)">
          <ZButton @click="edit">수정</ZButton>
          <ZButton @click="del">삭제</ZButton>
        </template>
      </div>
    </div>
  </div>
</template>
