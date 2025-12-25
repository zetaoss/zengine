<!-- ViewerReplies.vue -->
<script setup lang="ts">
import AvatarUser from '@common/components/avatar/AvatarUser.vue'
import ZButton from '@common/ui/ZButton.vue'
import ZIcon from '@common/ui/ZIcon.vue'
import ZMenu from '@common/ui/ZMenu.vue'
import ZMenuItem from '@common/ui/ZMenuItem.vue'
import ZModal from '@common/ui/ZModal.vue'
import ZTextarea from '@common/ui/ZTextarea.vue'
import httpy from '@common/utils/httpy'
import { mdiDotsVertical } from '@mdi/js'
import { useDateFormat } from '@vueuse/core'
import { computed, ref, watch } from 'vue'

import useAuthStore from '@/stores/auth'

import type { Reply } from '../types'
import BoxHTML from './ViewerHTML.vue'

const props = defineProps({
  postId: { type: Number, default: 0 },
})

const me = useAuthStore()

const replies = ref<Reply[]>([])
const replyBody = ref('')
const editingReply = ref<{ id: number; body: string } | null>(null)
const showModal = ref(false)
const deletingReply = ref<Reply | null>(null)

const isEditing = (id: number) => editingReply.value?.id === id

const editBody = computed({
  get: () => editingReply.value?.body ?? '',
  set: (v: string) => {
    if (!editingReply.value) return
    editingReply.value.body = v
  },
})

async function fetchData() {
  if (!props.postId) return
  const [data, err] = await httpy.get<Reply[]>(`/api/posts/${props.postId}/replies`)
  if (err) {
    console.error(err)
    return
  }
  replies.value = data
}

async function postReply() {
  if (!props.postId) return
  const [, err] = await httpy.post(`/api/posts/${props.postId}/replies`, {
    body: replyBody.value,
  })
  if (err) {
    console.error(err)
    return
  }

  replyBody.value = ''
  fetchData()
}

function edit(reply: Reply) {
  editingReply.value = { id: reply.id, body: reply.body }
}

async function editOK() {
  if (!props.postId || !editingReply.value) return

  const [, err] = await httpy.put(
    `/api/posts/${props.postId}/replies/${editingReply.value.id}`,
    { body: editingReply.value.body },
  )
  if (err) {
    console.error(err)
    return
  }

  editingReply.value = null
  fetchData()
}

function editCancel() {
  editingReply.value = null
}

function del(reply: Reply) {
  deletingReply.value = reply
  showModal.value = true
}

async function modalOK() {
  showModal.value = false
  if (!props.postId || !deletingReply.value) return

  const [, err] = await httpy.delete(`/api/posts/${props.postId}/replies/${deletingReply.value.id}`)
  if (err) {
    console.error(err)
    return
  }

  deletingReply.value = null
  fetchData()
}

watch(() => props.postId, fetchData, { immediate: true })
</script>

<template>
  <ZModal :show="showModal" @ok="modalOK" @cancel="showModal = false">
    댓글을 삭제하시겠습니까?
  </ZModal>

  <div>
    <h3 class="py-2 px-4 text-lg">
      댓글 ({{ replies.length }})
    </h3>

    <div v-for="reply in replies" :key="reply.id" class="border-b py-3 px-4 text-sm">
      <div class="grid grid-cols-2">
        <div>
          <AvatarUser :avatar="reply.avatar" />
          {{ useDateFormat(reply.created_at, 'YYYY-MM-DD HH:mm').value }}
        </div>

        <div v-if="reply.avatar && me.canEdit(reply.avatar.id)" class="text-right">
          <ZMenu>
            <template #trigger="{ toggle }">
              <button type="button" @click="toggle">
                <ZIcon :path="mdiDotsVertical" />
              </button>
            </template>

            <template #menu="{ close }">
              <div class="text-xs">
                <ZMenuItem @click="edit(reply); close()">
                  수정
                </ZMenuItem>
                <ZMenuItem @click="del(reply); close()">
                  삭제
                </ZMenuItem>
              </div>
            </template>
          </ZMenu>
        </div>
      </div>

      <div class="pt-2">
        <div v-if="isEditing(reply.id)">
          <div class="p-3 border-2 rounded bg-white dark:bg-black">
            <div class="flex items-center justify-between">
              <span class="text-xs text-gray-400">수정 중</span>
              <span class="text-xs text-gray-400">
                {{ editBody.length }} characters
              </span>
            </div>

            <div class="mt-2">
              <ZTextarea v-model="editBody" id="edit-reply" />
            </div>

            <div class="mt-3 flex justify-center gap-3">
              <ZButton :disabled="editBody.length === 0" @click="editOK" class="w-24" color="primary">
                저장
              </ZButton>
              <ZButton @click="editCancel" class="w-24">
                취소
              </ZButton>
            </div>
          </div>
        </div>

        <div v-else>
          <BoxHTML :body="reply.body" mode="text" :previews="false" :fencedCode="true" />
        </div>
      </div>
    </div>

    <div v-if="me.isLoggedIn && me.userData" class="p-3">
      <div class="p-4 border z-bg-muted rounded">
        <AvatarUser :avatar="me.userData.avatar" :showLink="false" />
        <ZTextarea v-model="replyBody" class="mt-2" id="new-reply" placeholder="댓글을 남겨보세요" />
        <div class="flex justify-end gap-3">
          <div class="text-xs text-gray-400">
            {{ replyBody.length }} 자
          </div>
          <ZButton :disabled="replyBody.length === 0" @click="postReply" class="w-20" color="primary">
            등록
          </ZButton>
        </div>
      </div>
    </div>
  </div>
</template>
