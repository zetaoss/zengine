<svelte:options customElement={{ tag: 'page-footer', shadow: 'none' }} />

<script lang="ts">
  import { mdiDotsHorizontal } from '@mdi/js'
  import { onMount } from 'svelte'

  import getRLCONF from '$lib/utils/rlconf'
  import AvatarIcon from '$shared/components/avatar/AvatarIcon.svelte'
  import CButton from '$shared/ui/CButton.svelte'
  import CMenu from '$shared/ui/CMenu.svelte'
  import CMenuItem from '$shared/ui/CMenuItem.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import httpy from '$shared/utils/httpy'
  import linkify from '$shared/utils/linkify'

  interface Row {
    id: number
    message: string
    messageHtml?: string
    user_id: number
    user_name: string
    created: string
    page_title: string
  }

  const { wgUserId, wgUserName, wgArticleId, wgUserGroups } = getRLCONF()

  let message = $state('')
  let docComments = $state<Row[]>([])
  let editingRow = $state<Row | null>(null)
  let deletingRow = $state<Row | null>(null)
  let showModal = $state(false)
  let fetchDataToken = 0

  const isLoggedIn = wgUserId > 0
  const isAdmin = (wgUserGroups || []).includes('sysop')

  const canEdit = (id: number) => isLoggedIn && wgUserId === id
  const canDelete = (id: number) => canEdit(id) || isAdmin

  async function fetchData() {
    const token = ++fetchDataToken
    const [data, err] = await httpy.get<Row[]>(`/api/comments/${wgArticleId}`)
    if (err) {
      console.log(err)
      return
    }

    const rows = data ?? []
    if (token !== fetchDataToken) return

    const linkedMessages = await linkify(rows.map((row) => row.message || ''))
    if (token !== fetchDataToken) return

    const nextDocComments = rows.map((row, i) => ({
      ...row,
      messageHtml: (linkedMessages[i] || '').replace(/\n/g, '<br />'),
    }))
    if (token !== fetchDataToken) return
    docComments = nextDocComments
  }

  async function postNew() {
    if (!message.trim()) return
    const [, err] = await httpy.post('/api/comments', {
      pageid: wgArticleId,
      message,
    })
    if (err) {
      console.log(err)
      return
    }
    message = ''
    fetchData()
  }

  function edit(row: Row) {
    editingRow = { ...row }
  }

  async function editOK() {
    if (!editingRow) return
    const [, err] = await httpy.put(`/api/comments/${editingRow.id}`, {
      message: editingRow.message,
    })
    if (err) {
      console.log(err)
      return
    }
    editingRow = null
    fetchData()
  }

  function editCancel() {
    editingRow = null
  }

  function del(row: Row) {
    deletingRow = row
    showModal = true
  }

  async function delOK() {
    showModal = false
    if (!deletingRow) return
    const [, err] = await httpy.delete(`/api/comments/${deletingRow.id}`)
    if (err) {
      console.log(err)
      return
    }
    deletingRow = null
    fetchData()
  }

  onMount(() => {
    fetchData()
  })
</script>

{#if showModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-foreground/40">
    <div class="bg-background rounded shadow-lg w-[22rem] max-w-[90vw] p-4 border">
      <div class="text-sm pb-3">댓글을 삭제하시겠습니까?</div>
      <div class="flex justify-end gap-2">
        <CButton variant="ghost" onclick={() => (showModal = false)}>취소</CButton>
        <CButton variant="destructive" onclick={delOK}>삭제</CButton>
      </div>
    </div>
  </div>
{/if}

<hr />

<div class="py-4">
  <div>문서 댓글 ({docComments.length})</div>

  {#if isLoggedIn}
    <div class="pt-3 flex">
      <div class="pt-1 pr-3">
        <AvatarIcon user={{ id: wgUserId, name: wgUserName }} size={32} />
      </div>

      <div class="w-full text-sm">
        <div>{wgUserName}</div>

        <div class="grid grid-cols-[5fr_1fr] gap-2">
          <textarea
            id="page-comment-new"
            class="w-full min-h-[4.5rem] p-2 border rounded bg-background"
            placeholder="댓글을 쓸 수 있습니다..."
            bind:value={message}
          ></textarea>
          <CButton variant="outline" class="h-full!" disabled={message.trim().length === 0} onclick={postNew}>등록</CButton>
        </div>
      </div>
    </div>
  {/if}

  {#each docComments as row (row.id)}
    <div class="mt-3 pt-2 flex border-t">
      <div class="pt-1 pr-3">
        <AvatarIcon user={{ id: row.user_id, name: row.user_name }} size={32} />
      </div>

      <div class="w-full">
        <div class="float-right flex items-center gap-1">
          {#if canEdit(row.user_id) || canDelete(row.user_id)}
            <CMenu>
              {#snippet trigger({ toggle })}
                <CButton variant="ghost" size="icon-sm" onclick={toggle}>
                  <ZIcon path={mdiDotsHorizontal} class="w-4 h-4 opacity-60" />
                </CButton>
              {/snippet}
              {#snippet menu({ close })}
                <div class="text-xs">
                  {#if canEdit(row.user_id)}
                    <CMenuItem
                      onclick={() => {
                        edit(row)
                        close()
                      }}>수정</CMenuItem
                    >
                  {/if}
                  {#if canDelete(row.user_id)}
                    <CMenuItem
                      onclick={() => {
                        del(row)
                        close()
                      }}>삭제</CMenuItem
                    >
                  {/if}
                </div>
              {/snippet}
            </CMenu>
          {/if}
        </div>

        <div class="text-sm">
          <a href={`/profile/${row.user_name}`}>{row.user_name}</a>
        </div>

        <!-- eslint-disable-next-line svelte/no-at-html-tags -->
        <div class="message text-sm">{@html row.messageHtml || ''}</div>
        <div class="text-sm text-muted-foreground">{row.created.substring(0, 10)}</div>

        {#if editingRow && row.id === editingRow.id}
          <div class="py-3">
            <div class="p-3 border-2 rounded">
              <div class="overflow-auto">
                <div class="float-left">
                  <AvatarIcon user={{ id: row.user_id, name: row.user_name }} size={32} />
                </div>
                <div class="float-right text-xs text-muted-foreground">
                  {editingRow.message.length} characters
                </div>
              </div>

              <div class="py-2">
                <textarea
                  id="page-comment-edit"
                  class="w-full min-h-[4.5rem] p-2 border rounded bg-background"
                  bind:value={editingRow.message}
                ></textarea>
              </div>

              <div class="overflow-auto">
                <div class="float-right flex gap-1">
                  <CButton variant="ghost" onclick={editCancel}>취소</CButton>
                  <CButton variant="default" disabled={editingRow.message.trim().length === 0} onclick={editOK}>저장</CButton>
                </div>
              </div>
            </div>
          </div>
        {/if}
      </div>
    </div>
  {/each}
</div>
