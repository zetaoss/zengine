<svelte:options runes={true} />

<script lang="ts">
  import { mdiAutoFix, mdiCreation, mdiDelete } from '@mdi/js'
  import { get } from 'svelte/store'

  import { goto } from '$app/navigation'
  import { resolve } from '$app/paths'
  import { page } from '$app/state'
  import ThePagination from '$lib/components/pagination/ThePagination.svelte'
  import type { PaginateData } from '$lib/components/pagination/types'
  import useAuthStore from '$lib/stores/auth'
  import AvatarUser from '$shared/components/avatar/AvatarUser.svelte'
  import EditBotModal from '$shared/components/editbot/EditBotModal.svelte'
  import { showConfirm } from '$shared/ui/confirm/confirm'
  import { showToast } from '$shared/ui/toast/toast'
  import ZButton from '$shared/ui/ZButton.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import ZTabs from '$shared/ui/ZTabs.svelte'
  import httpy from '$shared/utils/httpy'

  import WriteRequestNew from './WriteRequestNew.svelte'

  interface Row {
    hit: number
    id: number
    user_id: number
    user_name: string
    writer_id: number
    writer_name?: string
    rate: number
    ref: number
    title: string
    writed_at: string
    updated_at: string
  }

  interface RespData {
    current_page: number
    data: Row[]
    last_page: number
  }

  interface Count {
    done: number
    todo: number
  }

  interface EditBotTarget {
    title: string
    storeUrl: string
    requestType: 'create' | 'edit'
  }

  type Mode = 'todo' | 'todo-top' | 'done'

  const auth = useAuthStore()
  const canWrite = auth.canWrite
  const canDelete = auth.canDelete
  const userInfo = auth.userInfo

  let mode = $state<Mode>('todo')
  let respData = $state<RespData>({ current_page: 1, data: [], last_page: 1 })
  let paginateData = $state<PaginateData | null>(null)
  let currentPage = $state(1)
  let showModal = $state(false)
  let showEditBotModal = $state(false)
  let editBotRow = $state<Row | null>(null)
  let count = $state<Count>({ done: 0, todo: 0 })
  let loading = $state(true)

  let observedRoutePage = 0

  let routePage = $derived.by(() => {
    const p = Number(page.url.searchParams.get('page'))
    return Number.isFinite(p) && p > 0 ? p : 1
  })

  let isSysop = $derived(($userInfo?.groups ?? []).includes('sysop'))
  let modeTabs = $derived([
    { value: 'todo', label: '요청', badge: count.todo },
    { value: 'todo-top', label: '추천' },
    { value: 'done', label: '완료', badge: count.done },
  ])

  $effect(() => {
    if (routePage !== observedRoutePage) {
      observedRoutePage = routePage
      currentPage = routePage
      void fetchData()
    }
  })

  async function fetchCount() {
    const [data, err] = await httpy.get<Count>('/api/write-request/count')
    if (err) {
      console.error(err)
      return
    }
    count = data
  }

  async function fetchPage() {
    const [data, err] = await httpy.get<RespData>(`/api/write-request/${mode}`, {
      page: String(currentPage),
    })

    if (err) {
      console.error(err)
      loading = false
      return
    }

    respData = data
    paginateData = {
      current_page: data.current_page,
      last_page: data.last_page,
      path: '/tool/write-request',
    }

    loading = false
  }

  async function fetchData() {
    loading = true
    await fetchCount()
    if (typeof window !== 'undefined') {
      window.scrollTo(0, 0)
    }
    await fetchPage()
  }

  async function requireAuthConfirm(actionMsg: string) {
    if (get(canWrite)) return true

    if (
      await showConfirm(`${actionMsg}하려면 로그인이 필요합니다. 로그인하시겠습니까?`, {
        okText: '로그인',
      })
    ) {
      await goto(resolve(loginPath() as '/login'))
    }
    return false
  }

  async function openModal() {
    if (!(await requireAuthConfirm('작성요청을 등록'))) return

    showModal = true
  }

  function loginPath() {
    const searchParams = new URLSearchParams({
      returnto: `:${page.url.pathname}${page.url.search}`,
    })

    return `/login?${searchParams}`
  }

  function closeModal() {
    showModal = false
    void fetchData()
  }

  function setMode(nextMode: Mode) {
    mode = nextMode
    void fetchData()
  }

  async function del(row: Row) {
    const ok = await showConfirm(`'${row.title}' 건을 삭제하시겠습니까 ? `, { okColor: 'danger' })
    if (!ok) return

    const [, err] = await httpy.delete(`/api/write-request/${row.id}`)
    if (err) {
      console.error(err)
      showToast(err.message || '삭제 실패')
      return
    }

    await fetchData()
    showToast('삭제 완료')
  }

  async function addDocTask(row: Row) {
    if (!(await requireAuthConfirm('편집봇에 등록'))) return
    editBotRow = row
    showEditBotModal = true
  }

  function closeEditBotModal() {
    showEditBotModal = false
    editBotRow = null
  }

  async function recommend(row: Row) {
    if (!(await requireAuthConfirm('추천'))) return

    const [data, err] = await httpy.post<{ ok: boolean; rate: number }>(`/api/write-request/${row.id}/recommend`)
    if (err) {
      console.error(err)
      return
    }

    row.rate = data.rate
  }

  function getTitleHref(row: Row) {
    if (mode === 'done') return `/wiki/${row.title}`
    return `/w/index.php?search=${row.title}`
  }

  function getDateText(row: Row) {
    return row.updated_at.substring(0, 10)
  }

  function getWrittenDateText(row: Row) {
    const writedAt = (row.writed_at || '').trim()
    if (writedAt) return writedAt.substring(0, 10)
    return row.updated_at.substring(0, 10)
  }

  function getRequestUser(row: Row) {
    return { id: row.user_id, name: row.user_name }
  }

  function getDisplayUser(row: Row) {
    return {
      id: row.writer_id > 0 ? row.writer_id : 0,
      name: row.writer_name || (row.writer_id === 0 ? 'Unknown' : ''),
    }
  }

  let editBotTarget = $derived.by<EditBotTarget | null>(() => {
    if (!editBotRow) return null
    return {
      title: editBotRow.title,
      storeUrl: `/api/editbot/from-write-request/id/${editBotRow.id}`,
      requestType: mode === 'done' ? 'edit' : 'create',
    }
  })
</script>

<div class="p-5">
  <h2 class="my-5 text-2xl font-bold">작성 요청</h2>
  <WriteRequestNew show={showModal} on:close={closeModal} />
  <EditBotModal show={showEditBotModal} target={editBotTarget} onClose={closeEditBotModal} />

  <ZTabs tabs={modeTabs} selected={mode} onChange={(value) => setMode(value as Mode)} />

  <div class="flex justify-end py-3">
    <ZButton onclick={openModal}>등록</ZButton>
  </div>

  <table class="z-table">
    <thead>
      <tr>
        <th>번호</th>
        <th>제목</th>
        <th>추천</th>
        <th>검색</th>
        <th>역링크</th>
        {#if mode === 'done'}
          <th>요청</th>
          <th>작성</th>
        {:else}
          <th>요청자</th>
          <th>요청일</th>
        {/if}
        <th></th>
      </tr>
    </thead>
    {#if loading}
      <thead>
        <tr>
          <th colspan={8} class="p-0!">
            <div class="progress-wrap">
              <div class="progress-bar"></div>
            </div>
            {#if !respData.data.length}
              <div class="flex h-32 items-center justify-center">
                <ZSpinner />
              </div>
            {/if}
          </th>
        </tr>
      </thead>
    {/if}

    <tbody>
      {#each respData.data as row (row.id)}
        <tr>
          <td class="text-center">{row.id}</td>
          <td class="w-[35%]">
            <a href={getTitleHref(row)} rel="external" class={mode === 'done' ? '' : 'new'}>{row.title}</a>
          </td>
          <td class="text-center">
            {#if mode === 'done'}
              {row.rate}
            {:else}
              <ZButton class="min-w-10 px-2 py-1" onclick={() => recommend(row)}>{row.rate}</ZButton>
            {/if}
          </td>
          <td class="text-center">{row.hit}</td>
          <td class="text-center">
            <a href={`/wiki/특수:가리키는문서/${row.title}`} rel="external" class="btn">{row.ref}</a>
          </td>
          {#if mode === 'done'}
            <td class="text-center">
              <div>{getDateText(row)}</div>
              <div class="mt-1">
                <AvatarUser user={getRequestUser(row)} />
              </div>
            </td>
            <td class="text-center">
              <div>{getWrittenDateText(row)}</div>
              <div class="mt-1">
                <AvatarUser user={getDisplayUser(row)} />
              </div>
            </td>
          {:else}
            <td>
              <AvatarUser user={getRequestUser(row)} />
            </td>
            <td class="text-center">{getDateText(row)}</td>
          {/if}
          <td class="text-center">
            <div class="flex items-center justify-center gap-1">
              {#if isSysop}
                <ZButton color="default" size="small" title="편집봇에 추가" onclick={() => addDocTask(row)}>
                  <ZIcon path={mode === 'done' ? mdiAutoFix : mdiCreation} />
                </ZButton>
              {/if}
              {#if $canDelete(row.user_id)}
                <ZButton color="default" size="small" onclick={() => del(row)}>
                  <ZIcon path={mdiDelete} />
                </ZButton>
              {/if}
            </div>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>

  <div class="py-4 text-right">
    <ZButton onclick={openModal}>등록</ZButton>
  </div>

  {#if paginateData}
    <div class="pb-4">
      <ThePagination {paginateData} />
    </div>
  {/if}
</div>

<style>
  .progress-wrap {
    overflow: hidden;
    width: 100%;
    height: 4px;
    background: var(--background-color-progressive-subtle);
  }

  .progress-bar {
    width: 100%;
    height: 100%;
    animation: write-request-progress 1s infinite linear;
    background: var(--color-progressive);
    transform-origin: 0% 50%;
  }

  @keyframes write-request-progress {
    0% {
      transform: translateX(0) scaleX(0);
    }

    40% {
      transform: translateX(0) scaleX(0.4);
    }

    100% {
      transform: translateX(100%) scaleX(0.5);
    }
  }
</style>
