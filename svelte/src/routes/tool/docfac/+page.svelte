<svelte:options runes={true} />

<script lang="ts">
  import { mdiAutoFix, mdiCreation, mdiDelete, mdiPencil, mdiReplay } from '@mdi/js'
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
  import ZButton from '$shared/ui/ZButton.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import ZTabs from '$shared/ui/ZTabs.svelte'
  import httpy from '$shared/utils/httpy'
  import { getWikiEditHref } from '$shared/utils/wikiLink'

  interface DocTask {
    id: number
    user_id: number
    user_name: string
    title: string
    request_type: string
    status: string
    created_at: string
    updated_at: string
  }

  interface RespData {
    current_page: number
    data: DocTask[]
    last_page: number
  }

  interface QueueStatus {
    interval: string
    retry_interval: string
    retry_backoff: number
    phase: 'running' | 'sleeping' | 'retry_wait'
    next_run_at: string | null
    task_id: number | null
    retry_count: number
    last_error: string | null
    message: string
    next_run_after_seconds: number
    head: {
      id: number
      title: string
      status: string
      attempts: number
      error_count: number
      skip_count: number
      last_error: string | null
    } | null
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

  interface MwRawTextResp {
    query?: {
      pages?: MwPage[] | Record<string, MwPage>
    }
  }

  interface PromptDoc {
    label: string
    title: string
    content: string
    loading: boolean
    error: string | null
  }

  type Tab = 'list' | 'prompt'

  const tabs: Array<{ value: string; label: string }> = [
    { value: 'list', label: '목록' },
    { value: 'prompt', label: '프롬프트' },
  ]

  const promptDefs = {
    create: { label: '생성 프롬프트', title: '틀:문서공장 생성 프롬프트' },
    edit: { label: '편집 프롬프트', title: '틀:문서공장 편집 프롬프트' },
  }

  let rows = $state<DocTask[]>([])
  let loading = $state(true)
  let paginateData = $state<PaginateData | null>(null)
  let observedRoutePage = 0
  let deletingId = $state<number | null>(null)
  let cloningId = $state<number | null>(null)
  let tab = $state<Tab>('list')
  let queueStatus = $state<QueueStatus | null>(null)
  let runningNow = $state(false)
  let nowMs = $state(Date.now())
  let statusTimer: ReturnType<typeof setInterval> | null = null
  let prompts = $state<PromptDoc[]>(
    Object.values(promptDefs).map((doc) => ({
      ...doc,
      content: '',
      loading: false,
      error: null,
    })),
  )
  let promptsFetched = false

  const auth = useAuthStore()
  const userInfo = auth.userInfo

  let routePage = $derived.by(() => {
    const p = Number(page.url.searchParams.get('page'))
    return Number.isFinite(p) && p > 0 ? p : 1
  })

  let isSysop = $derived(($userInfo?.groups ?? []).includes('sysop'))
  let nextRunRemainingSeconds = $derived.by(() => {
    if (!queueStatus?.next_run_at) return 0
    const retryAt = new Date(queueStatus.next_run_at).getTime()
    if (Number.isNaN(retryAt)) return queueStatus.next_run_after_seconds
    return Math.max(0, Math.ceil((retryAt - nowMs) / 1000))
  })

  $effect(() => {
    if (routePage !== observedRoutePage) {
      observedRoutePage = routePage
      void fetchData()
      void fetchQueueStatus()
    }
  })

  $effect(() => {
    if (queueStatus?.next_run_at && !statusTimer) {
      statusTimer = setInterval(() => {
        nowMs = Date.now()
      }, 1000)
    }

    if (!queueStatus?.next_run_at && statusTimer) {
      clearInterval(statusTimer)
      statusTimer = null
    }
  })

  onDestroy(() => {
    if (statusTimer) {
      clearInterval(statusTimer)
      statusTimer = null
    }
  })

  $effect(() => {
    if (tab === 'prompt' && !promptsFetched) {
      promptsFetched = true
      void fetchPrompts()
    }
  })

  async function fetchData() {
    loading = true
    const [data, err] = await httpy.get<RespData>('/api/doctasks', {
      page: String(routePage),
    })
    loading = false

    if (err) {
      console.error(err)
      return
    }

    rows = data.data
    paginateData = {
      current_page: data.current_page,
      last_page: data.last_page,
      path: '/tool/docfac',
    }
  }

  async function fetchQueueStatus() {
    const [data, err] = await httpy.get<QueueStatus>('/api/doctasks/status')
    if (err) {
      console.error(err)
      return
    }

    queueStatus = data
  }

  function getUser(row: DocTask) {
    return { id: row.user_id, name: row.user_name }
  }

  function getRequestTypeLabel(requestType: string) {
    if (requestType === 'create') return '생성'
    if (requestType === 'edit') return '개선'
    return requestType || '-'
  }

  function getRequestTypeIcon(requestType: string) {
    if (requestType === 'create') return mdiCreation
    if (requestType === 'edit') return mdiAutoFix
    return null
  }

  function asArray<T>(value: T[] | Record<string, T> | undefined): T[] {
    if (!value) return []
    return Array.isArray(value) ? value : Object.values(value)
  }

  function getRevisionContent(pageData: MwPage | undefined) {
    const revision = pageData?.revisions?.[0]
    return revision?.slots?.main?.content ?? revision?.content ?? ''
  }

  async function fetchPrompt(doc: PromptDoc) {
    const [data, err] = await mwapi.get<MwRawTextResp>({
      action: 'query',
      prop: 'revisions',
      rvprop: 'content',
      rvslots: 'main',
      titles: doc.title,
    })

    if (err) {
      console.error(err)
      return { ...doc, loading: false, error: '불러오기 실패' }
    }

    const pageData = asArray(data?.query?.pages)[0]
    if (!pageData || pageData.missing) {
      return { ...doc, loading: false, error: '문서를 찾을 수 없습니다.' }
    }

    return { ...doc, content: getRevisionContent(pageData), loading: false, error: null }
  }

  async function fetchPrompts() {
    prompts = prompts.map((doc) => ({ ...doc, loading: true, error: null }))
    prompts = await Promise.all(prompts.map(fetchPrompt))
  }

  async function del(row: DocTask) {
    const ok = await showConfirm(`'${row.title}' 건을 삭제하시겠습니까 ? `)
    if (!ok) return

    deletingId = row.id
    const [, err] = await httpy.delete(`/api/doctasks/${row.id}`)
    deletingId = null

    if (err) {
      console.error(err)
      showToast(err.message || '삭제 실패')
      return
    }

    rows = rows.filter((item) => item.id !== row.id)
    void fetchQueueStatus()
    showToast('삭제 완료')
  }

  async function clone(row: DocTask) {
    const ok = await showConfirm(`'${row.title}' 건을 새 작업으로 등록하시겠습니까?`)
    if (!ok) return

    cloningId = row.id
    const [data, err] = await httpy.post<DocTask>(`/api/doctasks/${row.id}/clone`)
    cloningId = null

    if (err) {
      console.error(err)
      showToast(err.message || '복제 실패')
      return
    }

    rows = [data, ...rows]
    void fetchQueueStatus()
    showToast('복제해서 새 작업을 생성했습니다.')
  }

  async function runNowQueue() {
    if (runningNow) return

    runningNow = true
    const [data, err] = await httpy.post<QueueStatus>('/api/doctasks/run-now')
    runningNow = false

    if (err) {
      console.error(err)
      showToast(err.message || '바로 실행 실패')
      return
    }

    queueStatus = data
    nowMs = Date.now()
    showToast('지금 실행되도록 설정했습니다.')
  }

  function getQueueStatusClass(status: QueueStatus) {
    if (status.phase === 'retry_wait')
      return 'border-amber-300 bg-amber-50 text-amber-900 dark:border-amber-700 dark:bg-amber-950 dark:text-amber-100'
    if (status.phase === 'running')
      return 'border-blue-300 bg-blue-50 text-blue-900 dark:border-blue-700 dark:bg-blue-950 dark:text-blue-100'
    if (status.next_run_at) return 'border-gray-200 bg-gray-50 text-gray-700 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-200'
    if (status.head)
      return 'border-emerald-300 bg-emerald-50 text-emerald-900 dark:border-emerald-700 dark:bg-emerald-950 dark:text-emerald-100'
    return 'border-gray-200 bg-gray-50 text-gray-700 dark:border-gray-700 dark:bg-gray-900 dark:text-gray-200'
  }

  function formatCooldown(seconds: number) {
    const total = Math.max(0, Math.ceil(seconds))
    const hours = Math.floor(total / 3600)
    const minutes = Math.floor((total % 3600) / 60)
    const rest = total % 60
    if (hours > 0) {
      return `${hours}:${String(minutes).padStart(2, '0')}:${String(rest).padStart(2, '0')}`
    }
    return `${minutes}:${String(rest).padStart(2, '0')}`
  }

  function formatNextRunAt(value: string | null) {
    if (!value) return '-'
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return value

    const y = date.getFullYear()
    const m = String(date.getMonth() + 1).padStart(2, '0')
    const d = String(date.getDate()).padStart(2, '0')
    const hh = String(date.getHours()).padStart(2, '0')
    const mm = String(date.getMinutes()).padStart(2, '0')
    return `${y}-${m}-${d} ${hh}:${mm}`
  }

  function getPhaseText(phase: QueueStatus['phase']) {
    if (phase === 'running') return 'running'
    if (phase === 'retry_wait') return 'retry_wait'
    return 'sleeping'
  }
</script>

<div class="p-5">
  <h2 class="my-5 text-2xl font-bold">문서공장</h2>

  <ZTabs {tabs} selected={tab} onChange={(value) => (tab = value as Tab)} />

  {#if queueStatus}
    <div class={`mb-4 rounded border p-3 text-sm ${getQueueStatusClass(queueStatus)}`}>
      <div class="font-semibold">{getPhaseText(queueStatus.phase)}</div>
      <div class="mt-1">{queueStatus.message}</div>
      {#if isSysop}
        <div class="mt-2">
          <ZButton color="primary" size="small" disabled={runningNow} onclick={runNowQueue}>지금 실행</ZButton>
        </div>
      {/if}
      {#if queueStatus.head}
        <div class="mt-1">
          작업 예정:
          <a href={resolve(`/tool/docfac/${queueStatus.head.id}`)}>#{queueStatus.head.id} {queueStatus.head.title}</a>
          · 시도 {queueStatus.head.attempts}
          · 실패 {queueStatus.head.error_count}
          · skip {queueStatus.head.skip_count}
        </div>
      {/if}
      {#if queueStatus.next_run_at}
        <div class="mt-1 flex items-center gap-2">
          다음 실행 예정: {formatNextRunAt(queueStatus.next_run_at)}
          <span class="ml-1">({formatCooldown(nextRunRemainingSeconds)} 남음)</span>
        </div>
      {/if}
      {#if queueStatus.phase === 'retry_wait'}
        <div class="mt-1">재시도 {queueStatus.retry_count}</div>
        {#if queueStatus.last_error}
          <div class="mt-1 whitespace-pre-wrap text-xs opacity-80">{queueStatus.last_error}</div>
        {/if}
      {/if}
    </div>
  {/if}

  {#if tab === 'list'}
    <table class="mytable z-card w-full">
      <thead class="z-base3">
        <tr>
          <th>번호</th>
          <th>제목</th>
          <th>유형</th>
          <th>상태</th>
          <th>등록자</th>
          <th>등록일</th>
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
            <td colspan="6" class="text-center text-(--color-subtle)">문서공장 큐가 비어 있습니다.</td>
          </tr>
        {:else}
          {#each rows as row (row.id)}
            <tr class="border-b border-(--border-color-subtle)">
              <td class="px-2 text-center">{row.id}</td>
              <td>
                <a href={resolve(`/tool/docfac/${row.id}`)}>{row.title}</a>
                {#if isSysop}
                  <ZButton
                    color="ghost"
                    size="small"
                    class="align-middle leading-none text-(--color-subtle)"
                    disabled={deletingId === row.id}
                    onclick={() => del(row)}
                  >
                    <ZIcon path={mdiDelete} />
                  </ZButton>
                {/if}
              </td>
              <td>
                <div class="flex items-center justify-center gap-1">
                  {#if getRequestTypeIcon(row.request_type)}
                    <ZIcon path={getRequestTypeIcon(row.request_type) || ''} />
                  {/if}
                  <span>{getRequestTypeLabel(row.request_type)}</span>
                </div>
              </td>
              <td>
                <div class="flex items-center justify-center gap-1">
                  <span>{row.status}</span>
                  {#if isSysop && (row.status === 'completed' || row.status === 'retry_wait')}
                    <ZButton
                      color="ghost"
                      size="small"
                      class="leading-none text-(--color-subtle)"
                      disabled={cloningId === row.id}
                      onclick={() => clone(row)}
                    >
                      <ZIcon path={mdiReplay} />
                    </ZButton>
                  {/if}
                </div>
              </td>
              <td>
                <AvatarUser user={getUser(row)} />
              </td>
              <td class="text-center">{row.created_at.substring(0, 10)}</td>
            </tr>
          {/each}
        {/if}
      </tbody>
    </table>

    {#if paginateData}
      <div class="py-4">
        <ThePagination {paginateData} />
      </div>
    {/if}
  {:else}
    <div class="space-y-4">
      {#each prompts as prompt (prompt.title)}
        <section>
          <div class="mb-2 flex flex-wrap items-baseline gap-x-2 gap-y-1">
            <h3 class="text-lg font-semibold">{prompt.label}</h3>
            <a
              class="inline-flex text-(--color-subtle)"
              href={getWikiEditHref(prompt.title)}
              rel="external noopener noreferrer"
              target="_blank"
              title={`${prompt.title} 편집`}
              data-sveltekit-reload
            >
              <ZIcon path={mdiPencil} />
            </a>
          </div>

          {#if prompt.loading}
            <div class="flex h-24 items-center justify-center">
              <ZSpinner />
            </div>
          {:else if prompt.error}
            <div class="text-(--color-subtle)">{prompt.error}</div>
          {:else}
            <pre
              class="overflow-x-auto whitespace-pre-wrap rounded border border-(--border-color-subtle) bg-(--background-color-neutral-subtle) p-4 text-sm">{prompt.content ||
                '내용이 없습니다.'}</pre>
          {/if}
        </section>
      {/each}
    </div>
  {/if}
</div>

<style>
  th,
  td {
    padding: 0.5rem 1rem;
  }
</style>
