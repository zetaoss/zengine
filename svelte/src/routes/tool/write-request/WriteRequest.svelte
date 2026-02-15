<svelte:options runes={true} />

<script lang="ts">
  import { mdiDelete } from '@mdi/js'
  import { get } from 'svelte/store'

  import { page } from '$app/state'
  import ThePagination from '$lib/components/pagination/ThePagination.svelte'
  import type { PaginateData } from '$lib/components/pagination/types'
  import useAuthStore from '$lib/stores/auth'
  import AvatarUser from '$shared/components/avatar/AvatarUser.svelte'
  import { showConfirm } from '$shared/ui/confirm/confirm'
  import { showToast } from '$shared/ui/toast/toast'
  import ZButton from '$shared/ui/ZButton.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import httpy from '$shared/utils/httpy'

  import WriteRequestNew from './WriteRequestNew.svelte'

  interface Row {
    hit: number
    id: number
    user_id: number
    user_name: string
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

  type Mode = 'todo' | 'todo-top' | 'done'

  const auth = useAuthStore()
  const canWrite = auth.canWrite
  const canDelete = auth.canDelete

  let mode = $state<Mode>('todo')
  let respData = $state<RespData>({ current_page: 1, data: [], last_page: 1 })
  let paginateData = $state<PaginateData | null>(null)
  let currentPage = $state(1)
  let showModal = $state(false)
  let count = $state<Count>({ done: 0, todo: 0 })
  let loading = $state(true)

  let observedRoutePage = 0

  let routePage = $derived.by(() => {
    const p = Number(page.url.searchParams.get('page'))
    return Number.isFinite(p) && p > 0 ? p : 1
  })

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

  async function openModal() {
    if (!get(canWrite)) {
      window.location.href = '/login?redirect=/tool/write-request'
      return
    }
    showModal = true
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
    const ok = await showConfirm(`'${row.title}' 작성요청을 삭제하시겠습니까 ? `)
    if (!ok) return

    const [, err] = await httpy.delete(`/api/write-request/${row.id}`)
    if (err) {
      console.error(err)
      return
    }

    await fetchData()
    showToast('삭제 완료')
  }

  function getTitleHref(row: Row) {
    if (mode === 'done') return `/wiki/${row.title}`
    return `/w/index.php?search=${row.title}`
  }
</script>

<div class="p-5">
  <h2 class="my-5 text-2xl font-bold">작성 요청</h2>
  <WriteRequestNew show={showModal} on:close={closeModal} />

  <div class="pb-3">
    <button
      type="button"
      class={`inline-block rounded-l border p-3 ${mode === 'todo' ? 'bg-slate-100 dark:bg-slate-700' : 'bg-white dark:bg-slate-900'}`}
      onclick={() => setMode('todo')}
    >
      요청 <span class="rounded-full bg-gray-400 px-1 text-xs text-white">{count.todo}</span>
    </button>
    <button
      type="button"
      class={`inline-block border border-l-0 p-3 ${mode === 'todo-top' ? 'bg-slate-100 dark:bg-slate-700' : 'bg-white dark:bg-slate-900'}`}
      onclick={() => setMode('todo-top')}
    >
      추천
    </button>
    <button
      type="button"
      class={`inline-block rounded-r border border-l-0 p-3 ${mode === 'done' ? 'bg-slate-100 dark:bg-slate-700' : 'bg-white dark:bg-slate-900'}`}
      onclick={() => setMode('done')}
    >
      완료 <span class="rounded-full bg-gray-400 px-2 text-xs text-white">{count.done}</span>
    </button>
  </div>

  <table class="mytable z-card w-full">
    <thead class="z-base3">
      <tr>
        <th>번호</th>
        <th>제목</th>
        <th>추천</th>
        <th>검색</th>
        <th>역링크</th>
        <th>요청일</th>
        <th>요청자</th>
      </tr>
    </thead>

    {#if loading}
      <thead>
        <tr>
          <th colspan="9" class="p-0!">
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
        <tr class="align-top border-b border-[#88888866]">
          <td class="px-2 text-center">{row.id}</td>
          <td class="w-[35%]">
            <a href={getTitleHref(row)} rel="external" class={mode === 'done' ? '' : 'new'}>{row.title}</a>
            {#if $canDelete(row.user_id)}
              <ZButton color="ghost" class="py-1 align-middle leading-none text-[#888]" on:click={() => del(row)}>
                <ZIcon path={mdiDelete} />
              </ZButton>
            {/if}
          </td>
          <td class="text-center">{row.rate}</td>
          <td class="text-center">{row.hit}</td>
          <td class="text-center">
            <a href={`/wiki/특수:가리키는문서/${row.title}`} rel="external" class="btn">{row.ref}</a>
          </td>
          <td class="text-center">{row.updated_at.substring(0, 10)}</td>
          <td class="user">
            <AvatarUser user={{ id: row.user_id, name: row.user_name }} />
          </td>
        </tr>
      {/each}
    </tbody>
  </table>

  <div class="py-4 text-right">
    <ZButton disabled={!$canWrite} onclick={openModal}>등록</ZButton>
  </div>

  {#if paginateData}
    <div class="pb-4">
      <ThePagination {paginateData} />
    </div>
  {/if}
</div>

<style>
  th,
  td {
    padding: 0.5rem 1rem;
  }

  .progress-wrap {
    height: 4px;
    width: 100%;
    overflow: hidden;
    background: rgba(5, 114, 206, 0.05);
  }

  .progress-bar {
    height: 100%;
    width: 100%;
    animation: write-request-progress 1s infinite linear;
    transform-origin: 0% 50%;
    background: rgb(5, 114, 206);
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
