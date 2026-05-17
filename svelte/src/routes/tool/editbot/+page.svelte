<svelte:options runes={true} />

<script lang="ts">
  import { mdiCompare, mdiDelete, mdiEye, mdiPencil } from '@mdi/js'
  import { onDestroy } from 'svelte'

  import { resolve } from '$app/paths'
  import { page } from '$app/state'
  import ThePagination from '$lib/components/pagination/ThePagination.svelte'
  import type { PaginateData } from '$lib/components/pagination/types'
  import useAuthStore from '$lib/stores/auth'
  import mwapi from '$lib/utils/mwapi'
  import AvatarUser from '$shared/components/avatar/AvatarUser.svelte'
  import { showConfirm } from '$shared/ui/confirm/confirm'
  import { showToast } from '$shared/ui/toast/toast'
  import ZBadge from '$shared/ui/ZBadge.svelte'
  import ZButton from '$shared/ui/ZButton.svelte'
  import ZButtonLink from '$shared/ui/ZButtonLink.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import ZStatusText from '$shared/ui/ZStatusText.svelte'
  import ZTabs from '$shared/ui/ZTabs.svelte'
  import httpy from '$shared/utils/httpy'
  import { getWikiDiffHref, getWikiEditHref, getWikiViewHref } from '$shared/utils/wikiLink'

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

  interface MwRevision {
    content?: string
    slots?: {
      main?: {
        content?: string
      }
    }
  }

  interface MwPage {
    title: string
    missing?: boolean
    revisions?: MwRevision[]
  }

  interface MwAllPagesResp {
    query?: {
      allpages?: Array<{ pageid: number; ns: number; title: string }>
    }
  }

  interface PromptItem {
    title: string
  }

  type Tab = 'list' | 'prompt'

  const tabs: Array<{ value: string; label: string }> = [
    { value: 'list', label: '작업목록' },
    { value: 'prompt', label: '프롬프트' },
  ]

  let rows = $state<DocTask[]>([])
  let loading = $state(true)
  let paginateData = $state<PaginateData | null>(null)
  let observedRoutePage = 0
  let deletingId = $state<number | null>(null)
  let tab = $state<Tab>('list')
  let nowMs = $state(Date.now())
  let clockTimer: ReturnType<typeof setInterval> | null = null
  let refreshTimer: ReturnType<typeof setTimeout> | null = null
  let pollingActive = false
  let refreshDelayMs = 1000
  let promptList = $state<PromptItem[]>([])
  let selectedPromptTitle = $state('')
  let selectedPromptContent = $state('')
  let promptListLoading = $state(false)
  let promptContentLoading = $state(false)
  let promptListFetched = false
  const REFRESH_BACKOFF_FACTOR = 1.1

  const auth = useAuthStore()
  const userInfo = auth.userInfo

  let routePage = $derived.by(() => {
    const p = Number(page.url.searchParams.get('page'))
    return Number.isFinite(p) && p > 0 ? p : 1
  })

  let isSysop = $derived(($userInfo?.groups ?? []).includes('sysop'))
  let currentTask = $derived.by(
    () => rows.find((row) => ['Generating', 'Publishing', 'RetryingGenerate', 'RetryingPublish'].includes(row.phase)) ?? null,
  )
  $effect(() => {
    if (tab !== 'list') return
    if (routePage === observedRoutePage) return
    observedRoutePage = routePage
    refreshDelayMs = 1000
    void fetchData()
  })

  $effect(() => {
    if (tab === 'list') {
      startListPolling()
      return
    }
    stopListPolling()
  })

  $effect(() => {
    if (tab === 'prompt' && !promptListFetched) {
      promptListFetched = true
      void fetchPromptList()
    }
  })

  $effect(() => {
    const hasRetrying = rows.some((row) => row.phase === 'RetryingGenerate' || row.phase === 'RetryingPublish')
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
    const [data, err] = await httpy.get<RespData>('/api/editbot', {
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
      refreshDelayMs = 1000
    }
    paginateData = {
      current_page: data.current_page,
      last_page: data.last_page,
      path: '/tool/editbot',
    }
  }

  function startListPolling() {
    if (pollingActive) return
    pollingActive = true
    refreshDelayMs = 1000
    void fetchData()
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
    if (tab !== 'list' || !pollingActive) return
    if (refreshTimer) clearTimeout(refreshTimer)
    refreshTimer = setTimeout(async () => {
      if (tab !== 'list' || !pollingActive) return
      await fetchData({ silent: true })
      refreshDelayMs = Math.ceil(refreshDelayMs * REFRESH_BACKOFF_FACTOR)
      console.log('[editbot] refreshDelayMs increased to', refreshDelayMs)
      scheduleRefresh()
    }, refreshDelayMs)
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
    if (requestType === 'create') return 'text-emerald-600 dark:text-emerald-300'
    if (requestType === 'edit') return 'text-amber-600 dark:text-amber-300'
    return ''
  }

  function asArray<T>(value: T[] | Record<string, T> | undefined): T[] {
    if (!value) return []
    return Array.isArray(value) ? value : Object.values(value)
  }

  function getRevisionContent(pageData: MwPage | undefined) {
    const revision = pageData?.revisions?.[0]
    return revision?.slots?.main?.content ?? revision?.content ?? ''
  }

  interface MwRawTextResp {
    query?: {
      pages?: MwPage[] | Record<string, MwPage>
    }
  }

  async function fetchPromptContent(title: string) {
    if (selectedPromptTitle === title && selectedPromptContent) return
    selectedPromptTitle = title
    promptContentLoading = true
    selectedPromptContent = ''

    const [data, err] = await mwapi.get<MwRawTextResp>({
      action: 'query',
      prop: 'revisions',
      rvprop: 'content',
      rvslots: 'main',
      titles: title,
    })

    promptContentLoading = false
    if (err) {
      console.error(err)
      return
    }

    const pageData = asArray(data?.query?.pages)[0]
    if (!pageData || pageData.missing) {
      return
    }

    selectedPromptContent = getRevisionContent(pageData)
  }

  async function fetchPromptList() {
    promptListLoading = true
    const [data, err] = await mwapi.get<MwAllPagesResp>({
      action: 'query',
      list: 'allpages',
      apprefix: '프롬프트',
      apnamespace: 10,
      aplimit: 'max',
    })

    promptListLoading = false
    if (err) {
      console.error(err)
      return
    }

    promptList = (data?.query?.allpages ?? []).map((p) => ({ title: p.title }))
    if (promptList.length > 0) {
      void fetchPromptContent(promptList[0].title)
    }
  }

  async function del(row: DocTask) {
    const ok = await showConfirm(`'${row.title}' 건을 삭제하시겠습니까 ? `)
    if (!ok) return

    deletingId = row.id
    const [, err] = await httpy.delete(`/api/editbot/${row.id}`)
    deletingId = null

    if (err) {
      console.error(err)
      showToast(err.message || '삭제 실패')
      return
    }

    rows = rows.filter((item) => item.id !== row.id)
    refreshDelayMs = 1000
    void fetchData()
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
</script>

<div class="p-5">
  <h2 class="my-5 text-2xl font-bold">편집봇</h2>

  <ZTabs {tabs} selected={tab} onChange={(value) => (tab = value as Tab)} />

  {#if tab === 'list'}
    <div class="mb-3 flex justify-end">
      <ZButtonLink
        color="default"
        size="small"
        href="/wiki/특수:기여/Editbot"
        rel="external noopener noreferrer"
        target="_blank"
        data-sveltekit-reload
      >
        위키에서 보기
      </ZButtonLink>
    </div>
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
            <td colspan="6" class="text-center text-(--color-subtle)">편집봇 큐가 비어 있습니다.</td>
          </tr>
        {:else}
          {#each rows as row (row.id)}
            <tr class={isCurrentTask(row) ? 'bg-blue-50/70 dark:bg-blue-950/40' : ''}>
              <td class="text-center">{row.id}</td>
              <td>
                <ZBadge text={getRequestTypeLabel(row.request_type)} class={`mr-2 ${getRequestTypeClass(row.request_type)}`} />
                <a href={getWikiViewHref(row.title)} rel="external">{row.title}</a>
              </td>
              <td>
                <div class="flex items-center justify-center gap-1">
                  {#if ['Generating', 'Publishing'].includes(row.phase)}
                    <span class="animate-spin" aria-hidden="true">⌛</span>
                  {:else if ['Pending', 'RetryingGenerate', 'RetryingPublish'].includes(row.phase)}
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
                  <ZButtonLink color="default" size="small" href={resolve('/tool/editbot/[id]', { id: String(row.id) })}>
                    <ZIcon path={mdiEye} />
                  </ZButtonLink>
                  <ZButtonLink
                    color="default"
                    size="small"
                    href={getWikiDiffHref(row.title, row.revid)}
                    rel="external noopener noreferrer"
                    target="_blank"
                    title="미디어위키에서 차이보기"
                  >
                    <ZIcon path={mdiCompare} />
                  </ZButtonLink>
                  {#if isSysop}
                    <ZButton color="default" size="small" disabled={deletingId === row.id} onclick={() => del(row)}>
                      <ZIcon path={mdiDelete} />
                    </ZButton>
                  {/if}
                </div>
              </td>
            </tr>
            {#if isCurrentTask(row) && (row.error_count ?? 0) > 0}
              <tr class="border-b border-blue-200 bg-blue-50/50 dark:border-blue-800 dark:bg-blue-950/30">
                <td colspan="6" class="text-center text-sm">
                  <div class="text-blue-900 dark:text-blue-100">
                    실패 {row.error_count ?? 0}회
                    {#if row.last_error}
                      <small class="ml-1 opacity-80">{row.last_error}</small>
                    {/if}
                    {#if row.phase === 'RetryingGenerate' || row.phase === 'RetryingPublish'}
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
  {:else if tab === 'prompt'}
    <div class="flex flex-col gap-6 md:flex-row">
      <div class="flex flex-col gap-1 md:w-64 md:shrink-0">
        {#if promptListLoading}
          <div class="flex py-2">
            <ZSpinner />
          </div>
        {:else}
          {#each promptList as item (item.title)}
            <ZButton
              color={selectedPromptTitle === item.title ? 'default' : 'ghost'}
              size="small"
              class="justify-start text-left"
              onclick={() => void fetchPromptContent(item.title)}
            >
              {item.title.replace('틀:', '')}
            </ZButton>
          {/each}
        {/if}
      </div>

      <div class="min-w-0 flex-1">
        {#if selectedPromptTitle}
          <section>
            <div class="mb-2 flex items-baseline gap-2">
              <h3 class="text-lg font-semibold">{selectedPromptTitle}</h3>
              <ZButtonLink
                color="default"
                size="small"
                href={getWikiEditHref(selectedPromptTitle)}
                rel="external noopener noreferrer"
                target="_blank"
                title={`${selectedPromptTitle} 편집`}
                data-sveltekit-reload
              >
                <ZIcon path={mdiPencil} />
                <span>편집하기</span>
              </ZButtonLink>
            </div>

            {#if promptContentLoading}
              <div class="flex h-32 items-center justify-center">
                <ZSpinner />
              </div>
            {:else}
              <pre
                class="overflow-x-auto whitespace-pre-wrap rounded border border-(--border-color-subtle) bg-(--background-color-neutral-subtle) p-4 text-sm">{selectedPromptContent ||
                  '내용이 없습니다.'}</pre>
            {/if}
          </section>
        {:else if !promptListLoading}
          <div class="flex h-32 items-center justify-center text-(--color-subtle)">좌측에서 프롬프트를 선택해 주세요.</div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
</style>
