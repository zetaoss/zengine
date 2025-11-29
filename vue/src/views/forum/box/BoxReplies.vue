<script setup lang="ts">
import { ref, watch } from 'vue'

import { mdiDotsVertical } from '@mdi/js'
import { vOnClickOutside } from '@vueuse/components'
import { useDateFormat } from '@vueuse/core'

import ZIcon from '@common/ui/ZIcon.vue'
import ZModal from '@common/ui/ZModal.vue'
import ZTextarea from '@common/ui/ZTextarea.vue'
import ZButton from '@common/ui/ZButton.vue'
import UserAvatar from '@common/components/avatar/UserAvatar.vue'
import useAuthStore from '@/stores/auth'
import http from '@/utils/http'
import linkify from '@/utils/linkify'

import type { Reply } from '../types'

const props = defineProps({
  postID: { type: Number, default: 0 },
})

const me = useAuthStore()

const replies = ref([] as Reply[])
const replyBody = ref('')
const editingReply = ref({} as Reply)
const showModal = ref(false)
const dropdownReplyID = ref(0)

let deletingReply = {} as Reply

async function fetchData() {
  const resp = await http.get(`/api/posts/${props.postID}/replies`)
  resp.data.body = await linkify(resp.data.body)
  replies.value = resp.data
}

async function postReply() {
  await http.post(`/api/posts/${props.postID}/replies`, {
    body: replyBody.value,
  })
  replyBody.value = ''
  fetchData()
}

function edit(reply: Reply) {
  editingReply.value = { id: reply.id, body: reply.body } as Reply
}

async function editOK() {
  await http.put(`/api/posts/${props.postID}/replies/${editingReply.value.id}`, {
    body: editingReply.value.body,
  })
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
  try {
    await http.delete(`/api/posts/${props.postID}/replies/${deletingReply.id}`)
    fetchData()
  } catch (e) {
    console.error(e)
  }
}

function dropdown(reply: Reply) {
  dropdownReplyID.value = reply.id
}

function dropdownClose() {
  dropdownReplyID.value = 0
}

watch(() => props.postID, fetchData)
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
          <UserAvatar :user-avatar="reply.userAvatar" />
          {{ useDateFormat(reply.created_at, 'YYYY-MM-DD HH:mm').value }}
        </div>
        <div v-if="me && me.canEdit(reply.userAvatar.id)" v-on-click-outside="dropdownClose" class="text-right">
          <button type="button" @click="dropdown(reply)">
            <ZIcon :path="mdiDotsVertical" />
          </button>
        </div>
      </div>
      <div v-if="dropdownReplyID == reply.id" class="relative">
        <div class="z-10 absolute right-0 bg-white divide-y divide-gray-100 rounded shadow dark:bg-gray-700">
          <ul class="list-none p-0 text-xs text-gray-700 dark:text-gray-200">
            <li>
              <button type="button"
                class="block px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-600 dark:hover:text-white"
                @click="edit(reply)">
                수정
              </button>
            </li>
            <li>
              <button type="button"
                class="block px-3 py-2 hover:bg-gray-100 dark:hover:bg-gray-600 dark:hover:text-white"
                @click="del(reply)">
                삭제
              </button>
            </li>
          </ul>
        </div>
      </div>
      <div class="pt-2">
        <div class="whitespace-pre-line" v-html="reply.body" />
        <div v-if="editingReply && reply.id === editingReply.id" class="py-3">
          <div class="p-3 border-2 rounded bg-white dark:bg-black">
            <div class="overflow-auto">
              <div class="float-left">
                <UserAvatar :user-avatar="reply.userAvatar" :showLink="false" />
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
              <ZButton @click="editCancel">취소</ZButton>
              <ZButton :disabled="editingReply.body.length == 0" @click="editOK" color="primary">등록</ZButton>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div v-if="me.isLoggedIn" class="p-3">
      <div class="p-4 border z-bg-muted rounded">
        <UserAvatar :user-avatar="me.userData.avatar" :showLink="false" />
        <ZTextarea v-model="replyBody" class="mt-2" id="new-reply" placeholder="댓글을 남겨보세요" />
        <div class="flex justify-end gap-3">
          <div class="text-xs text-gray-400">{{ replyBody.length }} 자</div>
          <ZButton :disabled="replyBody.length == 0" @click="postReply" class="w-20" color="primary">등록</ZButton>
        </div>
      </div>
    </div>
  </div>
</template>
