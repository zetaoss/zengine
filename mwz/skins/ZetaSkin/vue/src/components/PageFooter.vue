<!-- PageFooter.vue -->
<script setup lang="ts">
import { ref } from 'vue'
import ZModal from '@common/ui/ZModal.vue'
import ZTextarea from '@common/ui/ZTextarea.vue'
import ZButton from '@common/ui/ZButton.vue'
import ZIcon from '@common/ui/ZIcon.vue'
import ZMenu from '@common/ui/ZMenu.vue'
import ZMenuItem from '@common/ui/ZMenuItem.vue'
import AvatarUser from '@common/components/avatar/AvatarUser.vue'
import type { Avatar } from '@common/components/avatar/avatar'
import { canDelete, canEdit } from '@/utils/auth'
import httpy from '@common/utils/httpy'
import getRLCONF from '@/utils/rlconf'
import LinkifyBox from './LinkifyBox.vue'
import { mdiDotsHorizontal } from '@mdi/js'

interface Row {
  avatar: Avatar
  id: number
  message: string
  created: string
  page_title: string
}

const { avatar, wgArticleId } = getRLCONF()

const message = ref('')
const showModal = ref(false)
const docComments = ref<Row[]>([])

const editingRow = ref<Row | null>(null)
const deletingRow = ref<Row | null>(null)

async function fetchData() {
  const [rows, err] = await httpy.get<Row[]>(`/api/comments/${wgArticleId}`)
  if (err) {
    console.log(err)
    return
  }
  docComments.value = rows
}

async function postNew() {
  if (!message.value) return

  const [, err] = await httpy.post('/api/comments', {
    pageid: wgArticleId,
    message: message.value,
  })
  if (err) {
    console.log(err)
    return
  }

  message.value = ''
  fetchData()
}

function edit(row: Row) {
  editingRow.value = { ...row }
}

async function editOK() {
  if (!editingRow.value) return

  const [, err] = await httpy.put(`/api/comments/${editingRow.value.id}`, {
    message: editingRow.value.message,
  })
  if (err) {
    console.log(err)
    return
  }

  editingRow.value = null
  fetchData()
}

function editCancel() {
  editingRow.value = null
}

function del(row: Row) {
  deletingRow.value = row
  showModal.value = true
}

async function delOK() {
  showModal.value = false
  const row = deletingRow.value
  if (!row) return

  const [, err] = await httpy.delete(`/api/comments/${row.id}`)
  if (err) {
    console.log(err)
    return
  }

  deletingRow.value = null
  fetchData()
}

fetchData()
</script>

<template>
  <ZModal :show="showModal" okColor="danger" @ok="delOK()" @cancel="showModal = false">
    댓글을 삭제하시겠습니까?
  </ZModal>

  <hr />

  <div class="p-6 py-4">
    <div>문서 댓글 ({{ docComments.length }})</div>

    <div class="pt-3 flex" v-if="avatar && avatar.id > 0">
      <div class="pt-1 pr-3">
        <AvatarUser :avatar="avatar" :showName="false" :showLink="false" :size="32" />
      </div>

      <div class="w-full text-sm">
        <div>{{ avatar.name }}</div>

        <div class="grid grid-cols-[5fr_1fr] gap-2">
          <ZTextarea v-model="message" id="page-comment-new" placeholder="댓글을 쓸 수 있습니다..." />
          <ZButton color="primary" :disabled="message.length === 0" @click="postNew">
            저장
          </ZButton>
        </div>
      </div>
    </div>

    <div class="mt-3 pt-2 flex border-t" v-for="row in docComments" :key="row.id">
      <div class="pt-1 pr-3">
        <AvatarUser :avatar="row.avatar" :showName="false" :showLink="false" :size="32" />
      </div>

      <div class="w-full">
        <div class="float-right">
          <ZMenu v-if="canEdit(row.avatar.id) || canDelete(row.avatar.id)">
            <template #trigger="{ toggle }">
              <ZButton size="small" color="ghost" aria-label="댓글 메뉴" @click="toggle()">
                <ZIcon :path="mdiDotsHorizontal" class="z-muted2" />
              </ZButton>
            </template>

            <template #menu="{ close }">
              <div class="text-xs">
                <ZMenuItem v-if="canEdit(row.avatar.id)" @click="edit(row); close()">
                  수정
                </ZMenuItem>
                <ZMenuItem v-if="canDelete(row.avatar.id)" @click="del(row); close()">
                  삭제
                </ZMenuItem>
              </div>
            </template>
          </ZMenu>
        </div>

        <div class="text-sm">
          <a :href="`/profile/${row.avatar.name}`">{{ row.avatar.name }}</a>
        </div>

        <LinkifyBox :content="row.message" />
        <div class="text-sm z-muted2">{{ row.created.substring(0, 10) }}</div>

        <div v-if="editingRow && row.id === editingRow.id" class="py-3">
          <div class="p-3 border-2 rounded">
            <div class="overflow-auto">
              <div class="float-left">
                <AvatarUser :avatar="row.avatar" />
              </div>
              <div class="float-right text-xs text-gray-400">
                {{ editingRow.message.length }} characters
              </div>
            </div>

            <div class="py-2">
              <ZTextarea v-model="editingRow.message" id="page-comment-edit" />
            </div>

            <div class="overflow-auto">
              <div class="float-right flex gap-1">
                <ZButton size="small" color="ghost" @click="editCancel">
                  취소
                </ZButton>
                <ZButton size="small" color="primary" :disabled="editingRow.message.length === 0" @click="editOK">
                  저장
                </ZButton>
              </div>
            </div>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>
