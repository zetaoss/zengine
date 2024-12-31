<script setup lang="ts">
import { ref } from 'vue'

import TheModal from '@common/components/TheModal.vue'
import TheTextarea from '@common/components/TheTextarea.vue'
import AvatarCore from '@common/components/avatar/AvatarCore.vue'
import AvatarUserLink from '@common/components/avatar/AvatarUserLink.vue'
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

interface CatComments {
  cats: {
    name: string
    cnt: number
  }[],
  comments: Row[],
}
const { avatar, wgArticleId } = getRLCONF()
const message = ref('')
const showModal = ref(false)
const docComments = ref([] as Row[])
const catComments = ref({} as CatComments)
const currentCat = ref('')

const editingRow = ref({} as Row)
let deletingRow = {} as Row

async function fetchDocComments() {
  docComments.value = await http.get(`/api/comments/${wgArticleId}`)
}

async function fetchCatComments() {
  catComments.value = await http.get(`/api/catcomments/${wgArticleId}`)
  if (catComments.value.cats.length < 1) return
  currentCat.value = catComments.value.cats[0].name
}

function fetchData() {
  fetchDocComments()
  fetchCatComments()
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

async function fetchCatCommentsCat(cat: string) {
  currentCat.value = cat
  const temp: CatComments = await http.get(`/api/catcomments-cat/${cat}`)
  catComments.value.comments = temp.comments
}

fetchData()
</script>
<template>
  <TheModal :show="showModal" ok-class="btn-danger" @ok="delOK()" @cancel="showModal = false">
    댓글을 삭제하시겠습니까?
  </TheModal>
  <div class="p-6 py-4">
    <div>
      문서 댓글 ({{ docComments.length }})
    </div>
    <div class="pt-3 flex" v-if="avatar && avatar.id > 0">
      <div class="pt-1 pr-3">
        <AvatarCore :user-avatar="avatar" :size="32" />
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
        <AvatarCore :userAvatar="row.avatar" :size="32" />
      </div>
      <div class="w-full">
        <button type="button" v-if="canDelete(row.avatar.id)" class="float-right btn btn-xs btn-danger" @click="del(row)">
          삭제
        </button>
        <button type="button" v-if="canEdit(row.avatar.id)" class="float-right btn btn-xs" @click="edit(row)">수정</button>
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
              <TheTextarea v-model="editingRow.message" />
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
  <div class="py-4">
    <div class="inline-block w-full border-b">
      <span class="pl-6 pr-2">분류 댓글:</span>
      <span v-for="cat in catComments.cats" :key="cat.name">
        <button type="button" :class="{ active: currentCat === cat.name }"
          class="p-2 px-3 rounded-t-lg hover:text-gray-600 hover:border-gray-300 dark:hover:text-gray-300"
          @click="fetchCatCommentsCat(cat.name)">
          {{ cat.name.replace(/_/g, ' ') }} <small>({{ cat.cnt }})</small>
        </button>
      </span>
    </div>
    <div class="p-6 md:columns-2">
      <div class="pt-3 w-full inline-block" v-for="row in catComments.comments" :key="row.id">
        <div class="text-sm">
          <a :href="'/wiki/' + row.page_title">
            {{ row.page_title.replace(/_/g, ' ') }}
          </a>
        </div>
        <div>
          <LinkifyBox :content="row.message" />
          ―
          <span class="text-green-600">
            <AvatarUserLink :userAvatar="row.avatar" />
          </span>
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
  @apply border border-b-transparent mb-[-3px] bg-z-card;
}
</style>
