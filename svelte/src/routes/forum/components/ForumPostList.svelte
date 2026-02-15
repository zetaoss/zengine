<svelte:options runes={true} />

<script lang="ts">
  import { resolve } from '$app/paths'
  import { page } from '$app/state'
  import ThePagination from '$lib/components/pagination/ThePagination.svelte'
  import type { PaginateData } from '$lib/components/pagination/types'
  import AvatarIcon from '$shared/components/avatar/AvatarIcon.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import httpy from '$shared/utils/httpy'

  import type { Post } from '../types'

  let { currentPostId = undefined, title = undefined }: { currentPostId?: number; title?: string } = $props()

  let posts = $state<Post[]>([])
  let paginateData = $state<PaginateData | null>(null)
  let isLoading = $state(false)
  let loadError = $state<string | null>(null)

  let pageNumber = $derived.by(() => {
    const p = Number(page.url.searchParams.get('page'))
    return Number.isFinite(p) && p > 0 ? p : 1
  })

  let lastPage = 0

  function formatDate(date: string) {
    const d = new Date(date)
    if (Number.isNaN(d.getTime())) return date
    const yy = String(d.getFullYear()).slice(-2)
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    const hh = String(d.getHours()).padStart(2, '0')
    const mi = String(d.getMinutes()).padStart(2, '0')
    return `${yy}-${mm}-${dd} ${hh}:${mi}`
  }

  async function fetchList() {
    isLoading = true
    loadError = null

    const [data, err] = await httpy.get<{
      data: Post[]
      current_page: number
      last_page: number
    }>('/api/posts', { page: pageNumber })

    if (err) {
      loadError = '목록을 불러오지 못했습니다.'
      posts = []
      paginateData = null
      isLoading = false
      return
    }

    posts = data.data
    paginateData = {
      current_page: data.current_page,
      last_page: data.last_page,
      path: '/forum',
    }
    isLoading = false
  }

  $effect(() => {
    if (pageNumber !== lastPage) {
      lastPage = pageNumber
      void fetchList()
    }
  })
</script>

<div class="text-sm">
  {#if title}
    <div class="mb-2 font-bold text-gray-700 dark:text-gray-300">
      {title}
    </div>
  {/if}

  <div class="z-card">
    <div class="z-base3 hidden p-2 text-center font-bold md:flex">
      <div class="flex w-[65%]">
        <span class="w-[10%]">번호</span>
        <span class="w-[90%] text-left">제목</span>
      </div>
      <div class="flex w-[35%]">
        <span class="w-[45%] text-left">작성자</span>
        <span class="w-[40%]">작성일</span>
        <span class="w-[15%]">조회</span>
      </div>
    </div>

    {#if isLoading}
      <div class="progress-wrap">
        <div class="progress-bar"></div>
      </div>
    {/if}

    {#if isLoading}
      <div class="flex items-center justify-center py-10 text-gray-500">
        <ZSpinner />
      </div>
    {:else if loadError}
      <div class="py-10 text-center text-red-500">
        {loadError}
      </div>
    {:else if posts.length === 0}
      <div class="py-10 text-center text-gray-500">아직 등록된 글이 없습니다.</div>
    {:else}
      {#each posts as p (p.id)}
        <a
          href={resolve(`/forum/${p.id}${pageNumber === 1 ? '' : `?page=${pageNumber}`}`)}
          class={`z-text block border-b px-3 py-2 hover:bg-gray-50 hover:no-underline dark:hover:bg-gray-800 md:flex md:px-2 ${
            currentPostId === p.id ? 'bg-slate-100 dark:bg-stone-900' : ''
          }`}
        >
          <div class="flex py-1 md:w-[65%]">
            <span class="hidden w-[10%] text-center md:inline">{p.id}</span>
            <span class="w-full truncate pr-2 md:w-[90%]">
              <span class="rounded-lg bg-[#6668] px-1.5 text-xs text-white dark:text-gray-200">
                {p.cat}
              </span>
              {p.title}
              {#if p.replies_count > 0}
                <span>({p.replies_count})</span>
              {/if}
            </span>
          </div>

          <div class="flex py-1 md:w-[35%]">
            <span class="flex-1 truncate md:w-[45%]">
              <AvatarIcon user={{ id: p.user_id, name: p.user_name }} size={15} />
              {p.user_name}
            </span>
            <span class="md:w-[40%] md:text-center">
              {formatDate(p.created_at)}
            </span>
            <span class="md:w-[15%] md:text-center">
              <span class="px-1 md:hidden">· 조회</span>
              {p.hit}
            </span>
          </div>
        </a>
      {/each}
    {/if}
  </div>

  {#if !isLoading && !loadError}
    <div class="py-4 text-center">
      {#if paginateData}
        <ThePagination {paginateData} />
      {/if}
    </div>
  {/if}
</div>

<style>
  .progress-wrap {
    height: 4px;
    width: 100%;
    overflow: hidden;
    background: rgba(127, 127, 127, 0.2);
  }

  .progress-bar {
    height: 100%;
    width: 35%;
    animation: forum-post-list-progress 1.2s ease-in-out infinite;
    background: #3b82f6;
  }

  @keyframes forum-post-list-progress {
    0% {
      transform: translateX(-120%);
    }
    100% {
      transform: translateX(380%);
    }
  }
</style>
