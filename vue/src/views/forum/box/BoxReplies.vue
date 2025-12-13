<!-- BoxRelies.vue -->
<script setup lang="ts">
import { ref, watch } from 'vue'

import { mdiDotsVertical } from '@mdi/js'
import { useDateFormat } from '@vueuse/core'

import ZIcon from '@common/ui/ZIcon.vue'
import ZModal from '@common/ui/ZModal.vue'
import ZTextarea from '@common/ui/ZTextarea.vue'
import ZButton from '@common/ui/ZButton.vue'
import ZMenu from '@common/ui/ZMenu.vue'
import ZMenuItem from '@common/ui/ZMenuItem.vue'
import AvatarUser from '@common/components/avatar/AvatarUser.vue'
import useAuthStore from '@/stores/auth'
import httpy from '@common/utils/httpy'
import linkify from '@/utils/linkify'

import type { Reply } from '../types'

const props = defineProps({
  postId: { type: Number, default: 0 },
})

const me = useAuthStore()

const replies = ref<Reply[]>([])
const replyBody = ref('')
const editingReply = ref({} as Reply)
const showModal = ref(false)

let deletingReply = {} as Reply

async function fetchData() {
  const [data, err] = await httpy.get<Reply[]>(`/api/posts/${props.postId}/replies`)
  if (err) {
    console.error(err)
    return
  }

  const linked = await Promise.all(
    data.map(async (r) => ({
      ...r,
      body: await linkify(r.body),
    })),
  )

  replies.value = linked
}

async function postReply() {
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
  editingReply.value = { id: reply.id, body: reply.body } as Reply
}

async function editOK() {
  const [, err] = await httpy.put(
    `/api/posts/${props.postId}/replies/${editingReply.value.id}`,
    { body: editingReply.value.body },
  )
  if (err) {
    console.error(err)
    return
  }

  editingReply.value = {} as Reply
  fetchData()
}

function editCancel() {
  editingReply.value = {} as Reply
}

function del(reply: Reply) {
  deletingReply = reply
  showModal.value = true
}

async function modalOK() {
  showModal.value = false
  const [, err] = await httpy.delete(`/api/posts/${props.postId}/replies/${deletingReply.id}`)
  if (err) {
    console.error(err)
    return
  }

  fetchData()
}

watch(() => props.postId, fetchData)
fetchData()
</script>

<template>
  <ZModal :show="showModal" @ok="modalOK()" @cancel="showModal = false">
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

        <div v-if="me && me.canEdit(reply.avatar.id)" class="text-right">
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
        <div class="whitespace-pre-line" v-html="reply.body" />

        <div v-if="editingReply && reply.id === editingReply.id" class="py-3">
          <div class="p-3 border-2 rounded bg-white dark:bg-black">
            <div class="overflow-auto">
              <div class="float-left">
                <AvatarUser :avatar="reply.avatar" :showLink="false" />
              </div>
              <div v-if="editingReply.body.length > 0" class="float-right">
                <div class="text-xs text-gray-400">
                  {{ editingReply.body.length }} characters
                </div>
              </div>
            </div>

            <div class="py-2 rounded-t bg-white dark:bg-black">
              <ZTextarea v-model="editingReply.body" id="edit-reply" />
            </div>

            <div class="flex justify-end gap-3">
              <ZButton @click="editCancel">
                취소
              </ZButton>
              <ZButton :disabled="editingReply.body.length === 0" @click="editOK" color="primary">
                등록
              </ZButton>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="me.isLoggedIn" class="p-3">
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
