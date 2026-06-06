<svelte:options runes={true} />

<script lang="ts">
  import { mdiCompare, mdiDelete, mdiInformation } from '@mdi/js'
  import { onDestroy } from 'svelte'

  import { resolve } from '$app/paths'
  import { page } from '$app/state'
  import ThePagination from '$lib/components/pagination/ThePagination.svelte'
  import type { PaginateData } from '$lib/components/pagination/types'
  import useAuthStore from '$lib/stores/auth'
  import AvatarUser from '$shared/components/avatar/AvatarUser.svelte'
  import { showConfirm } from '$shared/ui/confirm/confirm'
  import { showToast } from '$shared/ui/toast/toast'
  import ZBadge from '$shared/ui/ZBadge.svelte'
  import ZButton from '$shared/ui/ZButton.svelte'
  import ZButtonLink from '$shared/ui/ZButtonLink.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import ZStatusText from '$shared/ui/ZStatusText.svelte'
  import httpy from '$shared/utils/httpy'
  import { getWikiDiffHref, getWikiHref } from '$shared/utils/wikiLink'

  interface DocTask {
    id: number
    user_id: number
    user_name: string
    title: string
    request_type: string
    phase: string
    error_count?: number
    last_error?: string | null
    retry_at?: string | null
    llm_model?: string | null
    revid?: number
    created_at: string
    updated_at: string
  }

  interface RespData {
    current_page: number
    data: DocTask[]
    last_page: number
  }

  let rows = $state<DocTask[]>([])
  let loading = $state(true)
  let paginateData = $state<PaginateData | null>(null)
  let observedRoutePage = 0
  let deletingId = $state<number | null>(null)

  let nowMs = $state(Date.now())
  let clockTimer: ReturnType<typeof setInterval> | null = null
  let refreshTimer: ReturnType<typeof setTimeout> | null = null
  let pollingActive = false
  let refreshAfter = 1000
  const REFRESH_BACKOFF_FACTOR = 1.1
  let isFetching = false
  let isStale = false
  let staleNeedsLoud = false

  const auth = useAuthStore()
  const userInfo = auth.userInfo

  let routePage = $derived.by(() => {
    const p = Number(page.url.searchParams.get('page'))
    return Number.isFinite(p) && p > 0 ? p : 1
  })

  let isSysop = $derived(($userInfo?.groups ?? []).includes('sysop'))
  let currentTask = $derived.by(
    () => rows.find((row) => ['Generating', 'Publishing', 'Retrying'].includes(row.phase)) ?? null,
  )
  let hasActiveRows = $derived.by(() => rows.some((row) => !['Completed', 'Rejected', 'Failed'].includes(row.phase)))
  let activeCount = $derived.by(() => rows.filter((row) => !['Completed', 'Rejected', 'Failed'].includes(row.phase)).length)

  $effect(() => {
    if (routePage === observedRoutePage) return
    observedRoutePage = routePage
    refreshAfter = 1000
    void runFetchData()
  })

  $effect(() => {
    startListPolling()
    return () => {
      stopListPolling()
    }
  })

  $effect(() => {
    const hasRetrying = rows.some((row) => row.phase === 'Retrying')
    if (hasRetrying && !clockTimer) {
      clockTimer = setInterval(() => {
        nowMs = Date.now()
      }, 1000)
    }
    if (!hasRetrying && clockTimer) {
      clearInterval(clockTimer)
      clockTimer = null
    }
  })

  onDestroy(() => {
    if (clockTimer) clearInterval(clockTimer)
    stopListPolling()
  })

  async function fetchData(options: { silent?: boolean } = {}) {
    const prevSnapshot = rows.map((row) => `${row.id}:${row.phase}`).join('|')

    if (!options.silent) loading = true
    const [data, err] = await httpy.get<RespData>('/api/ai-edit', {
      page: String(routePage),
    })
    if (!options.silent) loading = false

    if (err) {
      console.error(err)
      return
    }

    rows = data.data
    const nextSnapshot = rows.map((row) => `${row.id}:${row.phase}`).join('|')
    if (prevSnapshot !== '' && prevSnapshot !== nextSnapshot) {
      refreshAfter = 1000
    }
    paginateData = {
      current_page: data.current_page,
      last_page: data.last_page,
      path: '/tool/ai-edit/tasks',
    }
  }

  async function runFetchData(options: { silent?: boolean } = {}): Promise<boolean> {
    const silent = options.silent ?? false
    if (isFetching) {
      isStale = true
      staleNeedsLoud = staleNeedsLoud || !silent
      return false
    }

    isFetching = true
    try {
      await fetchData({ silent })
    } finally {
      isFetching = false
    }

    let repeated = false
    if (isStale) {
      const nextSilent = !staleNeedsLoud
      isStale = false
      staleNeedsLoud = false
      repeated = true
      const subRepeated = await runFetchData({ silent: nextSilent })
      if (subRepeated) repeated = true
    }
    return repeated
  }

  function startListPolling() {
    if (pollingActive) return
    pollingActive = true
    refreshAfter = 1000
    void runFetchData()
    scheduleRefresh()
  }

  function stopListPolling() {
    pollingActive = false
    if (refreshTimer) {
      clearTimeout(refreshTimer)
      refreshTimer = null
    }
  }

  function scheduleRefresh() {
    if (!pollingActive) return
    if (!hasActiveRows) return
    if (refreshTimer) clearTimeout(refreshTimer)
    refreshTimer = setTimeout(async () => {
      if (!pollingActive) return
      await runFetchData({ silent: true })
      refreshAfter = Math.ceil(refreshAfter * REFRESH_BACKOFF_FACTOR)
      console.log('[aiedit] refreshAfter=%d active=%d', refreshAfter, activeCount)
      scheduleRefresh()
    }, refreshAfter)
  }

  function getUser(row: DocTask) {
    return { id: row.user_id, name: row.user_name }
  }

  function getRequestTypeLabel(requestType: string) {
    if (requestType === 'create') return '생성'
    if (requestType === 'edit') return '편집'
    return requestType || '-'
  }

  function getRequestTypeClass(requestType: string) {
    if (requestType === 'create') return 'text-emerald-600'
    if (requestType === 'edit') return 'text-amber-600'
    return ''
  }

  function getTitleHref(row: DocTask) {
    const exists = row.phase === 'Completed' || row.request_type !== 'create'
    return getWikiHref(row.title, exists)
  }

  async function del(row: DocTask) {
    const ok = await showConfirm(`'${row.title}' 건을 삭제하시겠습니까 ? `)
    if (!ok) return

    deletingId = row.id
    const [, err] = await httpy.delete(`/api/ai-edit/${row.id}`)
    deletingId = null

    if (err) {
      console.error(err)
      showToast(err.message || '삭제 실패')
      return
    }

    rows = rows.filter((item) => item.id !== row.id)
    refreshAfter = 1000
    void runFetchData()
    showToast('삭제 완료')
  }

  function formatRemaining(value: string | null | undefined) {
    if (!value) return '-'
    const ms = new Date(value).getTime() - nowMs
    if (Number.isNaN(ms)) return '-'
    const total = Math.max(0, Math.ceil(ms / 1000))
    const h = Math.floor(total / 3600)
    const m = Math.floor((total % 3600) / 60)
    const s = total % 60
    if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
    return `${m}:${String(s).padStart(2, '0')}`
  }

  function isCurrentTask(row: DocTask) {
    return currentTask?.id === row.id
  }

  function isCompleted(row: DocTask) {
    return row.phase === 'Completed'
  }

  function canViewDetail(row: DocTask) {
    return isSysop || row.user_id === $userInfo?.id
  }
</script>

<div class="p-5">
  <table class="z-table">
    <thead>
      <tr>
        <th>번호</th>
        <th>제목</th>
        <th>상태</th>
        <th>모델</th>
        <th>등록자</th>
        <th>등록일</th>
        <th></th>
      </tr>
    </thead>
    <tbody>
      {#if loading && rows.length === 0}
        <tr>
          <td colspan="6">
            <div class="flex h-32 items-center justify-center">
              <ZSpinner />
            </div>
          </td>
        </tr>
      {:else if rows.length === 0}
        <tr>
          <td colspan="6" class="text-center text-(--color-subtle)">AI 편집 큐가 비어 있습니다.</td>
        </tr>
      {:else}
        {#each rows as row (row.id)}
          <tr class={isCurrentTask(row) ? 'bg-blue-50/70' : ''}>
            <td class="text-center">{row.id}</td>
            <td>
              <ZBadge text={getRequestTypeLabel(row.request_type)} class={`mr-2 ${getRequestTypeClass(row.request_type)}`} />
              <a
                href={getTitleHref(row)}
                rel="noopener"
                target="_blank"
                class="font-medium hover:underline"
                class:new={!isCompleted(row)}
              >
                {row.title}
              </a>
              <span class="ml-1 inline-flex items-center gap-1 align-middle">
                {#if isCompleted(row)}
                  <ZButtonLink
                    color="ghost"
                    size="small"
                    href={getWikiDiffHref(row.title, row.revid)}
                    rel="external noopener noreferrer"
                    target="_blank"
                    title="차이보기"
                  >
                    <ZIcon path={mdiCompare} />
                  </ZButtonLink>
                {/if}
                {#if canViewDetail(row)}
                  <ZButtonLink color="ghost" size="small" href={resolve(`/tool/ai-edit/tasks/${row.id}`)} title="작업 상세 보기">
                    <ZIcon path={mdiInformation} />
                  </ZButtonLink>
                {/if}
              </span>
            </td>
            <td>
              <div class="flex items-center justify-center gap-1">
                {#if ['Generating', 'Publishing'].includes(row.phase)}
                  <span class="animate-spin" aria-hidden="true">⌛</span>
                {:else if ['Pending', 'Retrying'].includes(row.phase)}
                  <span aria-hidden="true">⌛</span>
                {/if}
                <ZStatusText text={row.phase} />
              </div>
            </td>
            <td class="text-center text-sm text-(--color-subtle)">
              {row.llm_model ?? '-'}
            </td>
            <td>
              <AvatarUser user={getUser(row)} />
            </td>
            <td class="text-center">{row.created_at.substring(0, 10)}</td>
            <td class="text-center">
              <div class="flex items-center justify-center gap-1">
                {#if isSysop}
                  <ZButton color="default" size="small" disabled={deletingId === row.id} onclick={() => del(row)}>
                    <ZIcon path={mdiDelete} />
                  </ZButton>
                {/if}
              </div>
            </td>
          </tr>
          {#if isCurrentTask(row) && (row.error_count ?? 0) > 0}
            <tr class="border-b border-blue-200 bg-blue-50/50">
              <td colspan="6" class="text-center text-sm">
                <div class="text-blue-900">
                  실패 {row.error_count ?? 0}회
                  {#if row.last_error}
                    <small class="ml-1 opacity-80">{row.last_error}</small>
                  {/if}
                  {#if row.phase === 'Retrying'}
                    <span class="ml-2">{formatRemaining(row.retry_at)} 후 재시도</span>
                  {/if}
                </div>
              </td>
            </tr>
          {/if}
        {/each}
      {/if}
    </tbody>
  </table>

  {#if paginateData}
    <div class="py-4">
      <ThePagination {paginateData} />
    </div>
  {/if}
</div>

<style>
  .z-table th,
  .z-table td {
    padding-left: 0.5rem;
    padding-right: 0.5rem;
  }
</style>
