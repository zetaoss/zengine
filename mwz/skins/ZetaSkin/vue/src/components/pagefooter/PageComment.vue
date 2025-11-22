<script setup lang="ts">
import { ref } from 'vue'
import BaseModal from '@common/ui/BaseModal.vue'
import UiTextarea from '@common/ui/UiTextarea.vue'
import AvatarUser from '@common/components/avatar/AvatarUser.vue'
import type UserAvatar from '@common/types/userAvatar'
import { canDelete, canEdit } from '@/utils/auth'
import http from '@/utils/http'
import getRLCONF from '@/utils/rlconf'
import LinkifyBox from '../LinkifyBox.vue'

interface Row {
  avatar: UserAvatar
  id: number
  message: string
  created: string
  page_title: string
}

const { avatar, wgArticleId } = getRLCONF()
const message = ref('')
const showModal = ref(false)
const docComments = ref([] as Row[])

const editingRow = ref({} as Row)
let deletingRow = {} as Row

async function fetchDocComments() {
  docComments.value = await http.get(`/api/comments/${wgArticleId}`)
}

function fetchData() {
  fetchDocComments()
}

async function postNew() {
  await http.post('/api/comments', {
    pageid: wgArticleId,
    message: message.value,
  })
  message.value = ''
  fetchData()
}

function edit(row: Row) {
  editingRow.value = row
}

async function editOK() {
  await http.put(`/api/comments/${editingRow.value.id}`, { message: editingRow.value.message })
  editingRow.value.id = 0
  fetchData()
}

function editCancel() {
  editingRow.value = {} as Row
}

function del(row: Row) {
  deletingRow = row
  showModal.value = true
}

async function delOK() {
  showModal.value = false
  await http.delete(`/api/comments/${deletingRow.id}`)
  fetchData()
}

fetchData()
</script>
<template>
  <BaseModal :show="showModal" ok-class="btn-danger" @ok="delOK()" @cancel="showModal = false">
    댓글을 삭제하시겠습니까?
  </BaseModal>
  <div class="p-6 py-4">
    <div>
      문서 댓글 ({{ docComments.length }})
    </div>
    <div class="pt-3 flex" v-if="avatar && avatar.id > 0">
      <div class="pt-1 pr-3">
        <AvatarUser :user-avatar="avatar" :showName="false" :showLink="false" :size="32" />
      </div>
      <div class="w-full text-sm">
        <div>{{ avatar.name }}</div>
        <div class="grid grid-cols-[5fr_1fr]">
          <textarea v-model="message" class="border rounded p-2" placeholder="댓글을 쓸 수 있습니다..." />
          <button type="button" class="btn btn-primary" @click="postNew">저장</button>
        </div>
      </div>
    </div>
    <div class="pt-3 flex" v-for="row in docComments" :key="row.id">
      <div class="pt-1 pr-3">
        <AvatarUser :userAvatar="row.avatar" :showName="false" :showLink="false" :size="32" />
      </div>
      <div class="w-full">
        <button type="button" v-if="canDelete(row.avatar.id)" class="float-right btn btn-xs btn-danger"
          @click="del(row)">
          삭제
        </button>
        <button type="button" v-if="canEdit(row.avatar.id)" class="float-right btn btn-xs"
          @click="edit(row)">수정</button>
        <div class="text-sm">
          <a :href="`/profile/${row.avatar.name}`">{{ row.avatar.name }}</a>
          <span class="ml-3 text-neutral-400">{{ row.created.substring(0, 10) }}</span>
        </div>
        <LinkifyBox :content="row.message" />
        <div v-if="editingRow && row.id === editingRow.id" class="py-3">
          <div class="p-3 border-2 rounded">
            <div class="overflow-auto">
              <div class="float-left">
                <AvatarUser :user-avatar="row.avatar" />
              </div>
              <div v-if="editingRow.message.length > 0" class="float-right">
                <div class="text-xs text-gray-400">
                  {{ editingRow.message.length }} characters
                </div>
              </div>
            </div>
            <div class="py-2 rounded-t">
              <UiTextarea v-model="editingRow.message" id="page-comment" />
            </div>
            <div class="overflow-auto">
              <div class="float-right">
                <button type="button" class="btn btn-xs" @click="editCancel">
                  취소
                </button>
                <button type="button" class="btn btn-xs btn-primary" :disabled="editingRow.message.length == 0"
                  @click="editOK">
                  저장
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
a.avatar {
  @apply text-gray-400;
}

.active {
  @apply border border-b-transparent mb-[-3px];
}
</style>
