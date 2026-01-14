<!-- @/views/forum/viewer/ViewerApex.vue -->
<script setup lang="ts">
import AvatarUser from '@common/components/avatar/AvatarUser.vue'
import ZButton from '@common/ui/ZButton.vue'
import ZModal from '@common/ui/ZModal.vue'
import ZSpinner from '@common/ui/ZSpinner.vue'
import httpy from '@common/utils/httpy'
import { useDateFormat } from '@vueuse/core'
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import useAuthStore from '@/stores/auth'
import RouterLinkButton from '@/ui/RouterLinkButton.vue'

import type { Post } from '../types'
import ViewerHTML from './ViewerHTML.vue'
import ViewerReplies from './ViewerReplies.vue'

const props = defineProps({ postId: Number })
const router = useRouter()
const auth = useAuthStore()

const post = ref<Post | null>(null)
const showModal = ref(false)
const isLoading = ref(false)

async function fetchData() {
  if (!props.postId) return
  isLoading.value = true
  try {
    const [data, err] = await httpy.get<Post>(`/api/posts/${props.postId}`)
    if (err) {
      console.error('Failed to load post:', err)
      post.value = null
      return
    }
    post.value = data
  } finally {
    isLoading.value = false
  }
}

function edit() {
  router.push({ path: `/forum/${props.postId}/edit` })
}

function del() {
  showModal.value = true
}

async function modalOK() {
  showModal.value = false
  const [, err] = await httpy.delete(`/api/posts/${props.postId}`)
  if (err) {
    console.error(err)
    return
  }
  window.location.href = '/forum'
}

watch(() => props.postId, fetchData, { immediate: true })
</script>

<template>
  <ZModal :show="showModal" @ok="modalOK" @cancel="showModal = false">
    글을 삭제하시겠습니까?
  </ZModal>

  <div class="border rounded py-4">
    <div class="px-4">
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

      <div v-if="post" class="py-3">
        <AvatarUser :user="{ id: post.user_id, name: post.user_name }" />
        <div class="text-sm">
          <span class="mr-4">
            {{ useDateFormat(post.created_at, 'YYYY-MM-DD HH:mm').value }}
          </span>
          <span>조회 {{ post.hit }}</span>
        </div>
      </div>

      <hr>

      <div v-if="isLoading" class="py-10 text-center text-gray-500 min-h-[9rem]">
        <ZSpinner />
      </div>

      <div v-else-if="post">
        <ViewerHTML class="py-4 min-h-[9rem]" :body="post.body" />
        <hr>
        <ViewerReplies :postId="postId" />
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
