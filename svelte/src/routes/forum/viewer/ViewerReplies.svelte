<script lang="ts">
  import { mdiDotsVertical } from '@mdi/js'

  import useAuthStore from '$lib/stores/auth'
  import AvatarUser from '$shared/components/avatar/AvatarUser.svelte'
  import ZButton from '$shared/ui/ZButton.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZMenu from '$shared/ui/ZMenu.svelte'
  import ZMenuItem from '$shared/ui/ZMenuItem.svelte'
  import ZModal from '$shared/ui/ZModal.svelte'
  import ZTextarea from '$shared/ui/ZTextarea.svelte'
  import httpy from '$shared/utils/httpy'

  import type { Reply } from '../types'
  import BoxHTML from './ViewerHTML.svelte'

  export let postId = 0

  const auth = useAuthStore()
  const { userInfo, isLoggedIn, canEdit } = auth

  let replies: Reply[] = []
  let replyBody = ''
  let editingReply: { id: number; body: string } | null = null
  let showModal = false
  let deletingReply: Reply | null = null

  $: editBody = editingReply?.body ?? ''

  function isEditing(id: number) {
    return editingReply?.id === id
  }

  function setEditBody(v: string) {
    if (!editingReply) return
    editingReply = { ...editingReply, body: v }
  }

  async function fetchData() {
    if (!postId) return
    const [data, err] = await httpy.get<Reply[]>(`/api/posts/${postId}/replies`)
    if (err) {
      console.error(err)
      return
    }
    replies = data
  }

  async function postReply() {
    if (!postId) return
    const [, err] = await httpy.post(`/api/posts/${postId}/replies`, {
      body: replyBody,
    })
    if (err) {
      console.error(err)
      return
    }

    replyBody = ''
    fetchData()
  }

  function edit(reply: Reply) {
    editingReply = { id: reply.id, body: reply.body }
  }

  async function editOK() {
    if (!postId || !editingReply) return

    const [, err] = await httpy.put(`/api/posts/${postId}/replies/${editingReply.id}`, { body: editingReply.body })
    if (err) {
      console.error(err)
      return
    }

    editingReply = null
    fetchData()
  }

  function editCancel() {
    editingReply = null
  }

  function del(reply: Reply) {
    deletingReply = reply
    showModal = true
  }

  async function modalOK() {
    showModal = false
    if (!postId || !deletingReply) return

    const [, err] = await httpy.delete(`/api/posts/${postId}/replies/${deletingReply.id}`)
    if (err) {
      console.error(err)
      return
    }

    deletingReply = null
    fetchData()
  }

  $: if (postId) fetchData()

  function formatDate(date: string) {
    const d = new Date(date)
    if (Number.isNaN(d.getTime())) return date
    const yyyy = d.getFullYear()
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    const hh = String(d.getHours()).padStart(2, '0')
    const mi = String(d.getMinutes()).padStart(2, '0')
    return `${yyyy}-${mm}-${dd} ${hh}:${mi}`
  }
</script>

<ZModal show={showModal} on:ok={modalOK} on:cancel={() => (showModal = false)}>댓글을 삭제하시겠습니까?</ZModal>

<div>
  <h3 class="pt-4 text-lg">댓글 ({replies.length})</h3>

  {#each replies as reply (reply.id)}
    <div class="border-b px-4 py-3 text-sm">
      <div class="grid grid-cols-2">
        <div>
          <AvatarUser user={{ id: reply.user_id, name: reply.user_name }} />
          <div class="text-xs text-gray-500">{formatDate(reply.created_at)}</div>
        </div>

        {#if $canEdit(reply.user_id)}
          <div class="text-right">
            <ZMenu>
              <svelte:fragment slot="trigger" let:toggle>
                <button type="button" on:click={toggle}>
                  <ZIcon path={mdiDotsVertical} />
                </button>
              </svelte:fragment>

              <svelte:fragment slot="menu" let:close>
                <div class="text-xs">
                  <ZMenuItem
                    on:click={() => {
                      edit(reply)
                      close()
                    }}
                  >
                    수정
                  </ZMenuItem>
                  <ZMenuItem
                    on:click={() => {
                      del(reply)
                      close()
                    }}
                  >
                    삭제
                  </ZMenuItem>
                </div>
              </svelte:fragment>
            </ZMenu>
          </div>
        {/if}
      </div>

      <div class="pt-2">
        {#if isEditing(reply.id)}
          <div class="rounded border-2 bg-white p-3 dark:bg-black">
            <div class="flex items-center justify-between">
              <span class="text-xs text-gray-400">수정 중</span>
              <span class="text-xs text-gray-400">{editBody.length} characters</span>
            </div>

            <div class="mt-2">
              <ZTextarea
                id="edit-reply"
                placeholder="댓글을 수정하세요"
                maxHeight={200}
                modelValue={editBody}
                onUpdateModelValue={setEditBody}
              />
            </div>

            <div class="mt-3 flex justify-center gap-3">
              <ZButton class="w-24" color="primary" disabled={editBody.length === 0} onclick={editOK}>저장</ZButton>
              <ZButton class="w-24" onclick={editCancel}>취소</ZButton>
            </div>
          </div>
        {:else}
          <BoxHTML body={reply.body} mode="text" previews={false} fencedCode={true} />
        {/if}
      </div>
    </div>
  {/each}

  {#if $isLoggedIn && $userInfo}
    <div class="p-3">
      <div class="z-bg-muted rounded border p-4">
        <AvatarUser user={$userInfo} showLink={false} />
        <div class="mt-2">
          <ZTextarea
            id="new-reply"
            placeholder="댓글을 남겨보세요"
            maxHeight={200}
            modelValue={replyBody}
            onUpdateModelValue={(value) => (replyBody = value)}
          />
        </div>
        <div class="flex justify-end gap-3">
          <div class="text-xs text-gray-400">{replyBody.length} 자</div>
          <ZButton class="w-20" color="primary" disabled={replyBody.length === 0} onclick={postReply}>등록</ZButton>
        </div>
      </div>
    </div>
  {/if}
</div>
