<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import { useDateFormat } from '@vueuse/core'
import Prism from 'prismjs'
import { useRouter } from 'vue-router'

import ZModal from '@common/ui/ZModal.vue'
import ZButton from '@common/ui/ZButton.vue'
import ZSpinner from '@common/ui/ZSpinner.vue'
import UserAvatar from '@common/components/avatar/UserAvatar.vue'
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
import RouterLinkButton from '@/ui/RouterLinkButton.vue'

const props = defineProps({ postID: Number })
const router = useRouter()
const auth = useAuthStore()

const post = ref<Post | null>(null)
const showModal = ref(false)
const isLoading = ref(false)

async function fetchData() {
  if (!props.postID) return
  isLoading.value = true
  try {
    const resp = await http.get(`/api/posts/${props.postID}`)
    post.value = resp.data
    await nextTick()
    Prism.highlightAll()
  } finally {
    isLoading.value = false
  }
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

watch(() => props.postID, fetchData, { immediate: true })
</script>

<template>
  <ZModal :show="showModal" @ok="modalOK" @cancel="showModal = false">
    글을 삭제하시겠습니까?
  </ZModal>

  <div class="border rounded py-4">
    <div class="px-4">
      <!-- 상단 메타 정보: post 있으면 내용, 없으면 비워두거나 스켈레톤 -->
      <div>
        <div class="float-left mr-2" v-if="post">
          <span class="text-xs p-1 rounded border text-gray-600 dark:text-gray-400">
            {{ post.cat }}
          </span>
        </div>
        <h3 class="text-xl">
          {{ post ? post.title : '' }}
        </h3>
      </div>

      <div v-if="post?.avatar" class="py-3">
        <UserAvatar :avatar="post.avatar" />
        <div class="text-sm">
          <span class="mr-4">
            {{ useDateFormat(post.created_at, 'YYYY-MM-DD HH:mm').value }}
          </span>
          <span>조회 {{ post.hit }}</span>
        </div>
      </div>

      <hr>

      <!-- ✨ 여기서부터 내용 영역만 로딩/본문 전환 -->
      <div v-if="isLoading" class="py-10 text-center text-gray-500 min-h-[9rem]">
        <ZSpinner />
      </div>

      <div v-else-if="post">
        <BoxHTML class="py-4 min-h-[9rem]" :body="post.body" />
        <hr>
        <BoxReplies :post-i-d="postID" />
      </div>

      <div v-else class="py-10 text-center text-gray-500 min-h-[9rem]">
        게시글을 불러올 수 없습니다.
      </div>
    </div>
  </div>

  <div class="py-4 flex gap-3">
    <RouterLinkButton :to="{ path: '/forum/new' }" :disabled="!auth.canWrite()">
      글쓰기
    </RouterLinkButton>

    <template v-if="post && auth.canEdit(post.user_id)">
      <ZButton @click="edit">수정</ZButton>
      <ZButton @click="del">삭제</ZButton>
    </template>
  </div>
</template>
