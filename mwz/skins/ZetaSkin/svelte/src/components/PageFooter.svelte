<svelte:options customElement={{ tag: 'page-footer', shadow: 'none' }} />

<script lang="ts">
  import { mdiDotsHorizontal } from '@mdi/js'
  import { onMount } from 'svelte'

  import getRLCONF from '$lib/utils/rlconf'
  import AvatarIcon from '$shared/components/avatar/AvatarIcon.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import httpy from '$shared/utils/httpy'
  import { titlesExist } from '$shared/utils/mediawiki'

  interface Row {
    id: number
    message: string
    user_id: number
    user_name: string
    created: string
    page_title: string
  }

  const { wgUserId, wgUserName, wgArticleId, wgUserGroups } = getRLCONF()

  let message = ''
  let docComments: Row[] = []
  let editingRow: Row | null = null
  let deletingRow: Row | null = null
  let showModal = false
  let titleExistsMap: Record<string, boolean> = {}
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
    docComments = rows

    const titles = [...new Set(rows.flatMap((row) => getWikiTitles(row.message)))]
    if (titles.length === 0) {
      titleExistsMap = {}
      return
    }

    const existsMap = await titlesExist(titles)
    if (token !== fetchDataToken) return
    titleExistsMap = existsMap
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

  const escapeHtml = (input: string) =>
    input.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;')

  const decodeEntities = (input: string) =>
    input
      .replace(/&amp;/g, '&')
      .replace(/&lt;/g, '<')
      .replace(/&gt;/g, '>')
      .replace(/&quot;/g, '"')
      .replace(/&#39;/g, "'")

  const getWikiTitles = (input: string): string[] => {
    const titles: string[] = []
    const re = /\[\[([^\]|]+)(?:\|[^\]]*)?\]\]/g
    let match: RegExpExecArray | null
    while ((match = re.exec(input || '')) !== null) {
      const title = decodeEntities(match[1] || '').trim()
      if (title.length > 0) titles.push(title)
    }
    return titles
  }

  const linkifyInline = (input: string) =>
    input.replace(/\[\[([^\]|]+)(?:\|([^\]]*))?\]\]|https?:\/\/[^\s<]+/g, (match, target, display) => {
      if (match.startsWith('[[')) {
        const rawTarget = decodeEntities(String(target)).trim()
        const exists = titleExistsMap[rawTarget] === true
        const classList = exists ? 'internal' : 'internal new'
        const href = `/wiki/${encodeURIComponent(rawTarget.replace(/ /g, '_'))}`
        return `<a href="${href}" class="${classList}" data-sveltekit-reload>${display || target}</a>`
      }
      const rawUrl = decodeEntities(match)
      const safeUrl = rawUrl.replace(/"/g, '%22')
      return `<a href="${safeUrl}" class="external" target="_blank" rel="noreferrer noopener">${match}</a>`
    })

  const formatMessage = (input: string) => {
    const escaped = escapeHtml(input || '')
    const linked = linkifyInline(escaped)
    return linked.replace(/\n/g, '<br />')
  }

  onMount(() => {
    fetchData()
  })
</script>

{#if showModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
    <div class="bg-white dark:bg-neutral-800 rounded shadow-lg w-[22rem] max-w-[90vw] p-4">
      <div class="text-sm pb-3">댓글을 삭제하시겠습니까?</div>
      <div class="flex justify-end gap-2">
        <button type="button" class="page-btn" on:click={() => (showModal = false)}> 취소 </button>
        <button type="button" class="page-btn bg-red-600 text-white" on:click={delOK}> 삭제 </button>
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
            class="w-full min-h-[4.5rem] p-2 border rounded bg-white dark:bg-neutral-900"
            placeholder="댓글을 쓸 수 있습니다..."
            bind:value={message}
          ></textarea>
          <button
            type="button"
            class="page-btn bg-blue-600 text-white disabled:opacity-50"
            disabled={message.trim().length === 0}
            on:click={postNew}
          >
            저장
          </button>
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
          {#if canEdit(row.user_id)}
            <button type="button" class="page-btn text-xs" on:click={() => edit(row)}> 수정 </button>
          {/if}
          {#if canDelete(row.user_id)}
            <button type="button" class="page-btn text-xs" on:click={() => del(row)}> 삭제 </button>
          {/if}
          {#if canEdit(row.user_id) || canDelete(row.user_id)}
            <ZIcon path={mdiDotsHorizontal} class="w-4 h-4 opacity-60" />
          {/if}
        </div>

        <div class="text-sm">
          <a href={`/profile/${row.user_name}`}>{row.user_name}</a>
        </div>

        <!-- eslint-disable-next-line svelte/no-at-html-tags -->
        <div class="message text-sm">{@html formatMessage(row.message)}</div>
        <div class="text-sm z-text3">{row.created.substring(0, 10)}</div>

        {#if editingRow && row.id === editingRow.id}
          <div class="py-3">
            <div class="p-3 border-2 rounded">
              <div class="overflow-auto">
                <div class="float-left">
                  <AvatarIcon user={{ id: row.user_id, name: row.user_name }} size={32} />
                </div>
                <div class="float-right text-xs text-gray-400">
                  {editingRow.message.length} characters
                </div>
              </div>

              <div class="py-2">
                <textarea
                  id="page-comment-edit"
                  class="w-full min-h-[4.5rem] p-2 border rounded bg-white dark:bg-neutral-900"
                  bind:value={editingRow.message}
                ></textarea>
              </div>

              <div class="overflow-auto">
                <div class="float-right flex gap-1">
                  <button type="button" class="page-btn text-xs" on:click={editCancel}> 취소 </button>
                  <button
                    type="button"
                    class="page-btn text-xs bg-blue-600 text-white disabled:opacity-50"
                    disabled={editingRow.message.trim().length === 0}
                    on:click={editOK}
                  >
                    저장
                  </button>
                </div>
              </div>
            </div>
          </div>
        {/if}
      </div>
    </div>
  {/each}
</div>
