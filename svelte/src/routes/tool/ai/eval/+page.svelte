<svelte:options runes={true} />

<script lang="ts">
  import { onDestroy, onMount } from 'svelte'

  import { mdiChevronDown, mdiInformationOutline } from '@mdi/js'

  import CButton from '$shared/ui/CButton.svelte'
  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZModal from '$shared/ui/ZModal.svelte'
  import httpy from '$shared/utils/httpy'
  import { displayTitle } from '$shared/utils/mediawiki'
  import { getAge } from '$shared/utils/time'
  import { getWikiViewHref } from '$shared/utils/wikiLink'

  import ModelId from '../ModelId.svelte'

  interface Criterion {
    id: string
    criterion: string
  }

  interface Rubrics {
    general_quality?: Criterion[]
    text_quality?: Criterion[]
  }

  interface CriterionVerdict {
    id: string
    reason: string
    verdict: string
  }

  interface Judgment {
    general_quality?: CriterionVerdict[]
    text_quality?: CriterionVerdict[]
    safety?: { reason?: string; violated?: boolean; policies?: string[] }
    accuracy?: {
      supported?: number
      contradictory?: number
      no_rad?: number
      errors?: Array<{ sentence: string; reason: string }>
    }
  }

  interface DocEvaluation {
    id: number
    page_id: number
    page_title: string
    revid: number
    definition: string
    grade: string
    general_quality: number
    text_quality: number
    score?: number
    safety_violated: boolean
    safety_policies: string
    accuracy_contradicted: number
    summary: string
    rubrics: string
    raw_json: string
    model: string
    prompt_tokens: number
    completion_tokens: number
    elapsed_ms: number
    created_at: string
    latest_revid?: number
    is_latest?: boolean
    revisions_after?: number
  }

  interface EvalListResp {
    total: number
    items: DocEvaluation[]
    limit: number
    offset: number
  }

  interface SummaryResp {
    total_evaluations: number
    avg_general_quality: number
    avg_text_quality: number
    grades: Array<{ grade: string; count: number }> | null
    contradicted: number
  }

  interface LastRun {
    at: string
    ok: boolean
    page_id?: number
    page_title?: string
    model?: string
    eval_id?: number
    error?: string
  }

  interface QueuedTask {
    id: string
    state: 'active' | 'pending' | 'scheduled' | 'failed'
    page_id?: number
    page_title?: string
    next_process_at?: string
    failed_at?: string
    last_error?: string
  }

  interface SettingsResp {
    interval: string
    model: string
    strategy?: string
    last_run?: LastRun | null
    tasks?: QueuedTask[] | null
    define_interval: string
    define_model: string
    define_strategy: string
    define_last_run?: LastRun | null
    define_tasks?: QueuedTask[] | null
    definitions: number
  }

  type SettingsChange = {
    interval?: string
    model?: string
    strategy?: string
    define_interval?: string
    define_model?: string
    define_strategy?: string
  }

  type SettingsSaved = Required<SettingsChange> & { ok: boolean }

  let loading = $state(true)
  let evaluations = $state<DocEvaluation[]>([])
  let totalCount = $state(0)
  let summary = $state<SummaryResp | null>(null)
  // A grade (S, A, B, C, D, F) or 'contradicted' for documents with contradicted sentences.
  let selectedFilter = $state<string>('')
  let expandedId = $state<number | null>(null)
  interface RevisionContent {
    loading: boolean
    error?: string
    text?: string
  }
  let revisionContents = $state<Record<number, RevisionContent>>({})
  let pageOffset = $state(0)
  const pageSize = 20

  let currentInterval = $state('1h')
  let currentModel = $state('groq/openai/gpt-oss-120b')
  let currentStrategy = $state('random')
  let currentDefineInterval = $state('30m')
  let currentDefineModel = $state('groq/openai/gpt-oss-120b')
  let currentDefineStrategy = $state('random')
  let definitionsCount = $state(0)
  let defineLastRun = $state<LastRun | null>(null)
  let defineTasks = $state<QueuedTask[]>([])
  let defineWaiting = $derived(defineTasks.filter((t) => t.state === 'pending' || t.state === 'scheduled'))
  let modelIds = $state<string[]>([])
  let modelOptions = $derived(currentModel && !modelIds.includes(currentModel) ? [currentModel, ...modelIds] : modelIds)
  let defineModelOptions = $derived(
    currentDefineModel && !modelIds.includes(currentDefineModel) ? [currentDefineModel, ...modelIds] : modelIds,
  )
  let savingSettings = $state(false)
  let settingsError = $state('')

  const gradeItems = [
    { key: 'S', color: 'text-a-amber-600' },
    { key: 'A', color: 'text-a-purple-600' },
    { key: 'B', color: 'text-a-blue-600' },
    { key: 'C', color: 'text-a-emerald-600' },
    { key: 'D', color: 'text-a-zinc-500' },
    { key: 'F', color: 'text-a-rose-600' },
  ]

  const policyLabels: Record<string, string> = {
    hate_harassment: '혐오·괴롭힘',
    sexual: '성적 콘텐츠',
    violence_gore: '폭력·잔혹',
    medical_advice: '의료 조언',
    dangerous: '위험 콘텐츠',
    csae: '아동 성착취',
    spam: '스팸',
    vandalism: '반달리즘',
  }

  function getGradeBadgeClass(grade: string): string {
    switch (grade) {
      case 'S':
        return 'bg-a-amber-100/20 text-a-amber-600 border-a-amber-600'
      case 'A':
        return 'bg-a-purple-100/20 text-a-purple-600 border-a-purple-600'
      case 'B':
        return 'bg-a-blue-100/20 text-a-blue-600 border-a-blue-600'
      case 'C':
        return 'bg-a-emerald-100/20 text-a-emerald-600 border-a-emerald-600'
      case 'D':
        return 'bg-a-zinc-100/20 text-a-zinc-500 border-a-zinc-500'
      case 'F':
        return 'bg-a-rose-100/20 text-a-rose-600 border-a-rose-600'
      default:
        return 'bg-a-zinc-100/20 text-a-zinc-500 border-a-zinc-500'
    }
  }

  function getScore(item: DocEvaluation): number {
    return item.score ?? Math.round(item.general_quality * 0.8 + item.text_quality * 0.2)
  }

  function criteriaWithVerdicts(criteria: Criterion[] | undefined, verdicts: CriterionVerdict[] | undefined) {
    return (criteria ?? []).map((c) => {
      const v = verdicts?.find((x) => x.id === c.id)
      return { ...c, pass: v?.verdict?.toLowerCase() === 'pass', reason: v?.reason ?? '' }
    })
  }

  function parseJSON<T>(raw: string): T | null {
    try {
      return JSON.parse(raw) as T
    } catch {
      return null
    }
  }

  function formatDate(iso: string): string {
    if (!iso) return ''
    const d = new Date(iso)
    if (isNaN(d.getTime())) return iso
    const pad = (n: number) => String(n).padStart(2, '0')
    const yyyy = d.getFullYear()
    const mm = pad(d.getMonth() + 1)
    const dd = pad(d.getDate())
    const hh = pad(d.getHours())
    const min = pad(d.getMinutes())
    return `${yyyy}-${mm}-${dd} ${hh}:${min}`
  }

  async function fetchSummary() {
    const [data, err] = await httpy.get<SummaryResp>('/api/evaluations/summary')
    if (err) {
      console.error('Failed to load evaluation summary:', err)
      return
    }
    summary = data
  }

  const intervalOptions = [
    { value: '0s', label: '끔' },
    { value: '10m', label: '10분' },
    { value: '30m', label: '30분' },
    { value: '1h', label: '1시간' },
    { value: '2h', label: '2시간' },
    { value: '6h', label: '6시간' },
    { value: '12h', label: '12시간' },
    { value: '24h', label: '24시간' },
  ]
  function withCurrent(value: string) {
    return intervalOptions.some((o) => o.value === value) ? intervalOptions : [...intervalOptions, { value, label: value }]
  }
  let selectOptions = $derived(withCurrent(currentInterval))
  let defineSelectOptions = $derived(withCurrent(currentDefineInterval))

  const strategyOptions = [
    { value: 'random', label: '랜덤' },
    { value: 'oldest', label: '오랜순' },
    { value: 'newest', label: '최신순' },
  ]
  const strategyLabels: Record<string, string> = {
    random: '랜덤',
    oldest: '오랜순',
    newest: '최신순',
  }

  let queuedTasks = $state<QueuedTask[]>([])
  let runningTasks = $derived(queuedTasks.filter((t) => t.state === 'active'))
  let waitingTasks = $derived(queuedTasks.filter((t) => t.state === 'pending' || t.state === 'scheduled'))
  const queueStateLabels: Record<QueuedTask['state'], string> = {
    active: '수행 중',
    pending: '대기',
    scheduled: '예약',
    failed: '실패',
  }
  let lastRun = $state<LastRun | null>(null)
  let now = $state(Date.now())
  let initialSettingsLoaded = false

  async function fetchSettings() {
    const [data, err] = await httpy.get<SettingsResp>('/api/evaluations/settings')
    if (!err && data) {
      currentInterval = data.interval
      currentModel = data.model
      if (data.strategy) currentStrategy = data.strategy
      queuedTasks = data.tasks ?? []
      applyDefineSettings(data)
      definitionsCount = data.definitions
      defineLastRun = data.define_last_run ?? null
      defineTasks = data.define_tasks ?? []
      const prevAt = lastRun?.at
      lastRun = data.last_run ?? null
      if (initialSettingsLoaded) {
        if (lastRun?.ok && lastRun.at !== prevAt) {
          fetchSummary()
          fetchEvaluations()
        }
      } else {
        initialSettingsLoaded = true
      }
    }
  }

  function formatTime(iso: string): string {
    const d = new Date(iso)
    if (isNaN(d.getTime())) return iso
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  }

  function formatCountdown(iso: string): string {
    const totalSecs = Math.max(0, Math.round((new Date(iso).getTime() - now) / 1000))
    const h = Math.floor(totalSecs / 3600)
    const m = Math.floor((totalSecs % 3600) / 60)
    const s = totalSecs % 60
    const pad = (n: number) => String(n).padStart(2, '0')
    if (h > 0) {
      return `${h}:${pad(m)}:${pad(s)}`
    }
    return `${pad(m)}:${pad(s)}`
  }

  function formatModelLabel(id: string): string {
    const i = id.indexOf('/')
    if (i === -1) return id
    return `${id.slice(0, i)} | ${id.slice(i + 1)}`
  }

  // Configured bob models in fallback order, for the model dropdown.
  async function fetchModelIds() {
    const [data, err] = await httpy.get<{
      usage: { data: Array<{ id: string }> } | null
      models: { data: Array<{ id: string; enabled: boolean }> }
    }>('/api/ai/models')
    if (err || !data) return
    modelIds = data.usage?.data.map((m) => m.id) ?? data.models.data.filter((m) => m.enabled).map((m) => m.id)
  }

  function applyDefineSettings(data: Pick<SettingsResp, 'define_interval' | 'define_model' | 'define_strategy'>) {
    currentDefineInterval = data.define_interval
    currentDefineModel = data.define_model
    currentDefineStrategy = data.define_strategy
  }

  async function handleSaveSettings(change: SettingsChange) {
    savingSettings = true
    settingsError = ''

    const [data, err] = await httpy.put<SettingsSaved>('/api/evaluations/settings', change)
    savingSettings = false

    if (err) {
      settingsError = err.message || '설정 저장에 실패했습니다.'
      return
    }

    if (data?.ok) {
      currentInterval = data.interval
      currentModel = data.model
      if (data.strategy) currentStrategy = data.strategy
      applyDefineSettings(data)
      showToast('저장했습니다')
    }
  }

  type CheckState = 'idle' | 'checking' | 'exists' | 'missing'

  interface SearchPage {
    id: number
    title: string
    key: string
    matched_title: string | null
    description: string | null
  }

  let triggeringRun = $state(false)
  let fetchingCandidate = $state(false)
  let showRunModal = $state(false)
  let targetTitle = $state('')
  let targetPageID = $state(0)
  let checkedTitle = $state('')
  let checkState = $state<CheckState>('idle')
  let checkSeq = 0
  let autoCheckTimer: ReturnType<typeof setTimeout> | undefined
  let titleInputEl = $state<HTMLInputElement | null>(null)

  let suggestions = $state<SearchPage[]>([])
  let showSuggestions = $state(false)
  let activeSuggestionIndex = $state(-1)
  let searchAborter: AbortController | null = null
  let searchDebounceTimer: ReturnType<typeof setTimeout> | null = null

  let trimmedTitle = $derived(targetTitle.trim())

  function clearAutoCheck() {
    if (autoCheckTimer) {
      clearTimeout(autoCheckTimer)
      autoCheckTimer = undefined
    }
  }

  function resetRunModal() {
    clearAutoCheck()
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    searchAborter?.abort()
    targetTitle = ''
    targetPageID = 0
    checkedTitle = ''
    checkState = 'idle'
    suggestions = []
    showSuggestions = false
    activeSuggestionIndex = -1
    checkSeq++
  }

  function openRunModal() {
    resetRunModal()
    showRunModal = true
    setTimeout(() => titleInputEl?.focus(), 50)
  }

  async function fetchSuggestions(q: string) {
    const trimmed = q.trim()
    if (!trimmed) {
      suggestions = []
      showSuggestions = false
      activeSuggestionIndex = -1
      return
    }

    searchAborter?.abort()
    const controller = new AbortController()
    searchAborter = controller

    try {
      const url = `/w/rest.php/v1/search/title?${new URLSearchParams({ q: trimmed, limit: '8' })}`
      const res = await fetch(url, { signal: controller.signal })
      if (!res.ok) return
      const json = (await res.json()) as { pages?: SearchPage[] }
      if (searchAborter === controller && targetTitle.trim() === trimmed) {
        suggestions = Array.isArray(json.pages) ? json.pages : []
        showSuggestions = suggestions.length > 0
        activeSuggestionIndex = -1
        const exact = suggestions.find((p) => p.title.toLowerCase() === trimmed.toLowerCase())
        if (exact) {
          checkState = 'exists'
          checkedTitle = trimmed
          targetPageID = exact.id
        }
      }
    } catch (e: unknown) {
      if (e instanceof DOMException && e.name === 'AbortError') return
    } finally {
      if (searchAborter === controller) searchAborter = null
    }
  }

  function debouncedFetchSuggestions(q: string) {
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    searchDebounceTimer = setTimeout(() => {
      void fetchSuggestions(q)
    }, 250)
  }

  function selectSuggestion(item: SearchPage) {
    clearAutoCheck()
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    searchAborter?.abort()

    targetTitle = item.title
    targetPageID = item.id
    checkedTitle = item.title
    checkState = 'exists'
    showSuggestions = false
    suggestions = []
    activeSuggestionIndex = -1
    checkSeq++
    setTimeout(() => titleInputEl?.focus(), 50)
  }

  async function checkDocumentExists(titleToCheck: string) {
    clearAutoCheck()
    const title = titleToCheck.trim()
    if (!title) {
      checkState = 'idle'
      checkedTitle = ''
      targetPageID = 0
      return
    }

    if (checkedTitle === title && checkState !== 'idle') {
      return
    }

    checkState = 'checking'
    const seq = ++checkSeq

    try {
      const [data, err] = await httpy.get<{
        query?: {
          pages?: Array<{ pageid?: number; title?: string; missing?: boolean }>
        }
      }>('/w/api.php', {
        action: 'query',
        format: 'json',
        formatversion: '2',
        titles: title,
      })
      if (seq !== checkSeq || targetTitle.trim() !== title) return
      checkedTitle = title
      const page = data?.query?.pages?.[0]
      if (!err && page && !page.missing && page.pageid && page.pageid > 0) {
        checkState = 'exists'
        targetPageID = page.pageid
        if (page.title) {
          targetTitle = page.title
          checkedTitle = page.title
        }
      } else {
        checkState = 'missing'
        targetPageID = 0
      }
    } catch {
      if (seq !== checkSeq || targetTitle.trim() !== title) return
      checkState = 'idle'
      targetPageID = 0
    }
  }

  function scheduleAutoCheck() {
    clearAutoCheck()
    const title = targetTitle.trim()
    if (!title) {
      checkState = 'idle'
      checkedTitle = ''
      return
    }
    if (checkedTitle === title && checkState !== 'idle') return

    autoCheckTimer = setTimeout(() => {
      void checkDocumentExists(title)
    }, 500)
  }

  function onTitleInput(e: Event) {
    targetTitle = (e.target as HTMLInputElement).value
    targetPageID = 0
    searchAborter?.abort()
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    checkSeq++
    checkState = 'idle'
    scheduleAutoCheck()
    debouncedFetchSuggestions(targetTitle)
  }

  function onTitleKeydown(e: KeyboardEvent) {
    if (e.isComposing) return

    if (e.key === 'ArrowDown') {
      if (showSuggestions && suggestions.length > 0) {
        e.preventDefault()
        activeSuggestionIndex = (activeSuggestionIndex + 1) % suggestions.length
        return
      }
    } else if (e.key === 'ArrowUp') {
      if (showSuggestions && suggestions.length > 0) {
        e.preventDefault()
        activeSuggestionIndex = (activeSuggestionIndex - 1 + suggestions.length) % suggestions.length
        return
      }
    } else if (e.key === 'Escape') {
      if (showSuggestions) {
        e.preventDefault()
        showSuggestions = false
        activeSuggestionIndex = -1
        return
      }
    } else if (e.key === 'Enter') {
      e.preventDefault()

      if (showSuggestions && activeSuggestionIndex >= 0 && activeSuggestionIndex < suggestions.length) {
        selectSuggestion(suggestions[activeSuggestionIndex])
        return
      }

      showSuggestions = false

      if (checkState === 'exists' && !triggeringRun) {
        void handleConfirmRun()
        return
      }

      if (targetTitle.trim() && checkState !== 'checking') {
        void checkDocumentExists(targetTitle.trim())
      }
    }
  }

  async function fetchAutoCandidate() {
    clearAutoCheck()
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    searchAborter?.abort()
    showSuggestions = false
    suggestions = []
    activeSuggestionIndex = -1

    fetchingCandidate = true
    const [data, err] = await httpy.get<{ page_id: number; page_title: string }>(
      `/api/evaluations/candidate?strategy=${encodeURIComponent(currentStrategy)}`,
    )
    fetchingCandidate = false

    if (err || !data?.page_title) {
      showToast('후보 문서를 가져오지 못했습니다.')
      return
    }

    targetTitle = data.page_title
    targetPageID = data.page_id
    checkedTitle = data.page_title
    checkSeq++
    checkState = 'exists'
    setTimeout(() => titleInputEl?.focus(), 50)
  }

  async function handleConfirmRun() {
    const titleToRun = targetTitle.trim()
    if (!titleToRun || checkState !== 'exists') return

    showRunModal = false
    triggeringRun = true
    const body = { page_id: targetPageID, page_title: titleToRun }
    const [, err] = await httpy.post<{ ok: boolean; task_id: string }>('/api/evaluations/run', body)
    triggeringRun = false

    if (err) {
      showToast(err.message || '평가 실행에 실패했습니다.')
      return
    }
    showToast(`'${titleToRun}' 평가를 시작했습니다`)
    resetRunModal()
    fetchSettings()
  }

  onDestroy(() => {
    clearAutoCheck()
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    searchAborter?.abort()
  })

  async function fetchEvaluations() {
    loading = true
    let query = `limit=${pageSize}&offset=${pageOffset}`
    if (selectedFilter === 'contradicted') {
      query += '&contradicted=1'
    } else if (selectedFilter) {
      query += `&grade=${encodeURIComponent(selectedFilter)}`
    }

    const [data, err] = await httpy.get<EvalListResp>(`/api/evaluations?${query}`)
    loading = false

    if (err) {
      console.error('Failed to load evaluations:', err)
      evaluations = []
      totalCount = 0
      return
    }

    if (data) {
      evaluations = data.items ?? []
      totalCount = data.total ?? 0
    }
  }

  function handleFilter(filter: string) {
    selectedFilter = filter
    pageOffset = 0
    fetchEvaluations()
  }

  function splitAtBytes(text: string, maxBytes: number): { evaluated: string; remaining: string } {
    const encoder = new TextEncoder()
    const bytes = encoder.encode(text)
    if (bytes.length <= maxBytes) {
      return { evaluated: text, remaining: '' }
    }
    const decoder = new TextDecoder('utf-8')
    return {
      evaluated: decoder.decode(bytes.slice(0, maxBytes)),
      remaining: decoder.decode(bytes.slice(maxBytes)),
    }
  }

  async function fetchRevisionContent(revid: number) {
    if (!revid || revid <= 0) return
    if (revisionContents[revid]?.text || revisionContents[revid]?.loading) return

    revisionContents[revid] = { loading: true }
    try {
      const resp = await fetch(`/w/rest.php/v1/revision/${revid}`)
      if (resp.ok) {
        const data = await resp.json()
        if (typeof data.source === 'string') {
          revisionContents[revid] = { loading: false, text: data.source }
          return
        }
      }

      const queryResp = await fetch(`/w/api.php?action=query&prop=revisions&revids=${revid}&rvprop=content&rvslots=main&format=json`)
      if (queryResp.ok) {
        const queryData = await queryResp.json()
        const pages = queryData.query?.pages
        if (pages) {
          const page = Object.values(pages)[0] as { revisions?: Array<{ slots?: { main?: { '*'?: string; content?: string } } }> }
          const src = page?.revisions?.[0]?.slots?.main?.['*'] ?? page?.revisions?.[0]?.slots?.main?.content
          if (typeof src === 'string') {
            revisionContents[revid] = { loading: false, text: src }
            return
          }
        }
      }

      revisionContents[revid] = { loading: false, error: '리비전 내용을 불러오지 못했습니다.' }
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : '네트워크 오류가 발생했습니다.'
      revisionContents[revid] = { loading: false, error: msg }
    }
  }

  function toggleExpand(item: DocEvaluation) {
    if (expandedId === item.id) {
      expandedId = null
    } else {
      expandedId = item.id
      if (item.revid > 0) {
        fetchRevisionContent(item.revid)
      }
    }
  }

  onMount(() => {
    fetchSummary()
    fetchSettings()
    fetchModelIds()
    fetchEvaluations()
    let tick = 0
    const timer = setInterval(() => {
      now = Date.now()
      tick++
      if (tick % 5 === 0) {
        fetchSettings()
      }
    }, 1000)
    return () => clearInterval(timer)
  })
</script>

<div class="container mx-auto p-4 md:p-6 max-w-6xl space-y-6">
  <div class="flex flex-wrap items-center justify-between gap-4 border-b pb-4">
    <h1 class="text-2xl font-bold tracking-tight">문서 평가</h1>
    <div class="flex flex-wrap items-center gap-2">
      <select
        class="h-8 max-w-64 rounded border bg-background px-2.5 text-xs focus:outline-none focus:ring-1 focus:ring-primary"
        value={currentModel}
        disabled={savingSettings}
        aria-label="모델"
        title="모델"
        onchange={(e) => handleSaveSettings({ model: e.currentTarget.value })}
      >
        {#each modelOptions as id (id)}
          <option value={id}>{formatModelLabel(id)}</option>
        {/each}
      </select>
      <select
        class="h-8 w-24 rounded border bg-background px-2.5 text-xs focus:outline-none focus:ring-1 focus:ring-primary"
        value={currentInterval}
        disabled={savingSettings}
        aria-label="주기"
        title="주기"
        onchange={(e) => handleSaveSettings({ interval: e.currentTarget.value })}
      >
        {#each selectOptions as option (option.value)}
          <option value={option.value}>{option.label}</option>
        {/each}
      </select>
      <select
        class="h-8 w-24 rounded border bg-background px-2.5 text-xs focus:outline-none focus:ring-1 focus:ring-primary"
        value={currentStrategy}
        disabled={savingSettings}
        aria-label="자동선정 방식"
        title="자동선정 방식"
        onchange={(e) => handleSaveSettings({ strategy: e.currentTarget.value })}
      >
        {#each strategyOptions as option (option.value)}
          <option value={option.value}>{option.label}</option>
        {/each}
      </select>
      <button
        type="button"
        class="h-8 rounded border bg-background px-3 text-xs font-medium text-foreground transition hover:bg-muted disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
        disabled={triggeringRun || runningTasks.length > 0}
        title="평가 즉시 실행 대화상자 열기"
        onclick={openRunModal}
      >
        Instant
      </button>
    </div>
  </div>

  <!-- 문서 정의: 평가는 지금 판의 정의가 있는 문서만 고른다 -->
  <div class="flex flex-wrap items-center justify-between gap-2 rounded-lg border bg-muted/30 p-3 text-xs text-muted-foreground">
    <div class="flex flex-wrap items-center gap-x-3 gap-y-1 min-w-0">
      <span class="font-medium text-foreground">문서 정의</span>
      <span>{definitionsCount.toLocaleString()}개</span>
      <span class="truncate">
        최근
        {#if defineLastRun}
          {getAge(defineLastRun.at, now)}
          {#if defineLastRun.page_title}
            <a
              href={(defineLastRun.page_id ?? 0) > 0
                ? `/w/index.php?curid=${defineLastRun.page_id}`
                : getWikiViewHref(defineLastRun.page_title)}
              target="_blank"
              rel="noopener">{displayTitle(defineLastRun.page_title)}</a
            >
          {/if}
          {#if defineLastRun.ok}
            <span class="font-medium text-a-emerald-600" title="성공">✓</span>
          {:else}
            <span class="font-medium text-a-rose-600" title={defineLastRun.error ?? '실패'}>✗</span>
          {/if}
        {:else}
          <span class="text-muted-foreground/60">-</span>
        {/if}
      </span>
      <span>
        다음
        {#if defineWaiting.length > 0 && defineWaiting[0].next_process_at}
          <span title={`${formatTime(defineWaiting[0].next_process_at)} 예정`}>{formatCountdown(defineWaiting[0].next_process_at)}</span>
        {:else}
          {currentDefineInterval === '0s' ? '꺼짐' : '없음'}
        {/if}
      </span>
    </div>
    <div class="flex flex-wrap items-center gap-2">
      <select
        class="h-8 max-w-64 rounded border bg-background px-2.5 text-xs focus:outline-none focus:ring-1 focus:ring-primary"
        value={currentDefineModel}
        disabled={savingSettings}
        aria-label="정의 모델"
        title="정의 모델"
        onchange={(e) => handleSaveSettings({ define_model: e.currentTarget.value })}
      >
        {#each defineModelOptions as id (id)}
          <option value={id}>{formatModelLabel(id)}</option>
        {/each}
      </select>
      <select
        class="h-8 w-24 rounded border bg-background px-2.5 text-xs focus:outline-none focus:ring-1 focus:ring-primary"
        value={currentDefineInterval}
        disabled={savingSettings}
        aria-label="정의 주기"
        title="정의 주기"
        onchange={(e) => handleSaveSettings({ define_interval: e.currentTarget.value })}
      >
        {#each defineSelectOptions as option (option.value)}
          <option value={option.value}>{option.label}</option>
        {/each}
      </select>
      <select
        class="h-8 w-24 rounded border bg-background px-2.5 text-xs focus:outline-none focus:ring-1 focus:ring-primary"
        value={currentDefineStrategy}
        disabled={savingSettings}
        aria-label="정의 자동선정 방식"
        title="정의 자동선정 방식"
        onchange={(e) => handleSaveSettings({ define_strategy: e.currentTarget.value })}
      >
        {#each strategyOptions as option (option.value)}
          <option value={option.value}>{option.label}</option>
        {/each}
      </select>
    </div>
  </div>
  {#if defineLastRun && !defineLastRun.ok && defineLastRun.error}
    <div class="-mt-4 px-3 text-xs text-a-rose-500 break-words">{defineLastRun.error}</div>
  {/if}

  <!-- 실행 상태 요약 바 -->
  <div class="flex flex-wrap items-start justify-between gap-4 rounded-lg border bg-muted/30 p-3 text-xs text-muted-foreground">
    <!-- 최근 -->
    <div class="space-y-0.5 min-w-0">
      <div class="font-medium text-foreground">
        최근 {#if lastRun}<span class="font-normal text-muted-foreground">{getAge(lastRun.at, now)}</span>{:else}<span
            class="font-normal text-muted-foreground/60">-</span
          >{/if}
      </div>
      <div class="truncate">
        {#if lastRun?.page_title}
          <a
            href={(lastRun.page_id ?? 0) > 0 ? `/w/index.php?curid=${lastRun.page_id}` : getWikiViewHref(lastRun.page_title)}
            target="_blank"
            rel="noopener"
          >
            {displayTitle(lastRun.page_title)}
          </a>
        {:else}
          <span class="text-muted-foreground/60">-</span>
        {/if}
      </div>
      <div class="truncate">
        {#if lastRun?.model}
          <span>{formatModelLabel(lastRun.model)}</span>
        {:else}
          <span class="text-muted-foreground/60">-</span>
        {/if}
        {#if lastRun}
          {#if lastRun.ok}
            <span class="ml-1 font-medium text-a-emerald-600" title="성공">✓</span>
          {:else}
            <span class="ml-1 font-medium text-a-rose-600" title="실패">✗</span>
          {/if}
        {/if}
      </div>
      {#if lastRun && !lastRun.ok && lastRun.error}
        <div class="text-a-rose-500 break-words">{lastRun.error}</div>
      {/if}
    </div>

    <!-- 실행 중 (있을 때만 표시) -->
    {#if runningTasks.length > 0}
      <div class="space-y-0.5 min-w-0">
        <div class="font-medium text-foreground flex items-center gap-1.5">
          <span>실행 중</span>
          <span class="h-2 w-2 rounded-full bg-a-blue-500 animate-pulse"></span>
        </div>
        <div class="truncate">
          {#if runningTasks[0].page_title}
            <a
              href={runningTasks[0].page_id && runningTasks[0].page_id > 0
                ? `/w/index.php?curid=${runningTasks[0].page_id}`
                : getWikiViewHref(runningTasks[0].page_title)}
              target="_blank"
              rel="noopener"
            >
              {displayTitle(runningTasks[0].page_title)}
            </a>
          {:else}
            <span class="text-muted-foreground">{strategyLabels[currentStrategy] ?? '랜덤'}</span>
          {/if}
        </div>
        <div class="truncate">
          <span>{formatModelLabel(currentModel)}</span>
        </div>
      </div>
    {/if}

    <!-- 다음 -->
    <div class="space-y-0.5 min-w-0">
      <div class="font-medium text-foreground">
        다음
        {#if waitingTasks.length === 0}
          <span class="font-normal text-muted-foreground">{currentInterval === '0s' ? '꺼짐' : '없음'}</span>
        {:else}
          {@const task = waitingTasks[0]}
          {#if task.next_process_at}
            <span class="font-normal text-muted-foreground" title={`${formatTime(task.next_process_at)} 예정`}
              >{formatCountdown(task.next_process_at)}</span
            >
          {:else}
            <span class="font-normal text-muted-foreground">{queueStateLabels[task.state]}</span>
          {/if}
        {/if}
      </div>
      <div class="truncate">
        {#if waitingTasks.length > 0}
          {#if waitingTasks[0].page_title}
            <a
              href={waitingTasks[0].page_id && waitingTasks[0].page_id > 0
                ? `/w/index.php?curid=${waitingTasks[0].page_id}`
                : getWikiViewHref(waitingTasks[0].page_title)}
              target="_blank"
              rel="noopener"
            >
              {displayTitle(waitingTasks[0].page_title)}
            </a>
          {:else}
            <span class="text-muted-foreground">{strategyLabels[currentStrategy] ?? '랜덤'}</span>
          {/if}
        {:else}
          <span class="text-muted-foreground/60">-</span>
        {/if}
      </div>
      <div class="truncate">
        {#if waitingTasks.length > 0 && currentInterval !== '0s'}
          <span>{formatModelLabel(currentModel)}</span>
        {:else}
          <span class="text-muted-foreground/60">-</span>
        {/if}
      </div>
    </div>
  </div>

  {#if settingsError}
    <div class="text-xs font-medium text-a-rose-600">
      ⚠️ {settingsError}
    </div>
  {/if}

  <!-- 등급별 필터 카드 -->
  <div class="grid grid-cols-4 sm:grid-cols-8 gap-2">
    {#each [{ key: '', label: '전체', color: 'text-foreground', count: summary?.total_evaluations ?? 0 }, ...gradeItems.map( (g) => ({ ...g, label: g.key, count: summary?.grades?.find((v) => v.grade === g.key)?.count ?? 0 }), ), { key: 'contradicted', label: '⚠ 오류', color: 'text-a-orange-600', count: summary?.contradicted ?? 0 }] as item (item.key)}
      <button
        type="button"
        class={`text-left rounded-lg border border-border p-3 shadow-sm transition ${
          selectedFilter === item.key ? 'bg-a-zinc-200' : 'bg-card hover:bg-a-zinc-100'
        }`}
        onclick={() => handleFilter(selectedFilter === item.key ? '' : item.key)}
      >
        <div class="text-xs font-medium text-muted-foreground">{item.label}</div>
        <div class={`text-xl font-bold mt-1 ${item.color}`}>{item.count === 0 ? '-' : item.count}</div>
      </button>
    {/each}
  </div>

  <!-- 목록 테이블 -->
  <div class="rounded-lg border bg-card overflow-hidden shadow-sm">
    <div class="overflow-x-auto">
      <table class="w-full text-left border-collapse text-sm">
        <thead>
          <tr class="border-b bg-muted/50 text-muted-foreground text-xs font-medium">
            <th class="p-3 whitespace-nowrap">일시</th>
            <th class="p-3">제목</th>
            <th class="p-3 text-center whitespace-nowrap">등급</th>
            <th class="p-3 text-center whitespace-nowrap">
              <div class="inline-flex items-center justify-center gap-1">
                <span>점수</span>
                <div class="relative group/tooltip inline-flex items-center">
                  <span
                    class="text-muted-foreground/70 hover:text-foreground cursor-help p-0.5 rounded"
                    aria-label="점수 및 등급 산정 방식 안내"
                  >
                    <ZIcon path={mdiInformationOutline} size={14} />
                  </span>
                  <div
                    class="absolute left-1/2 -translate-x-1/2 top-full mt-1.5 hidden group-hover/tooltip:block z-50 w-72 p-3 rounded-lg border border-border bg-popover text-popover-foreground text-xs shadow-lg font-normal leading-relaxed text-left pointer-events-none"
                  >
                    <div class="font-semibold text-foreground mb-1">점수 및 등급 산정 방식</div>
                    <div class="space-y-1.5 text-muted-foreground text-[11px]">
                      <div>
                        • <span class="text-foreground font-medium">점수 계산식</span>:
                        <div class="mt-0.5 pl-2 font-mono text-[10.5px] text-foreground/90">전반적 품질 80% + 텍스트 품질 20%</div>
                        <div class="pl-2 text-[10.5px] text-muted-foreground/80 space-y-0.5 mt-0.5">
                          <div>- 전반적 품질: 핵심 내용 기준 충족율</div>
                          <div>- 텍스트 품질: 문장력·문법·가독성 충족율</div>
                        </div>
                      </div>
                      <div>• <span class="text-foreground font-medium">등급 구간</span>:</div>
                      <div class="pl-2 space-y-0.5">
                        <div>S: 90점 이상 / A: 75~89점</div>
                        <div>B: 60~74점 / C: 40~59점</div>
                        <div>D: 40점 미만</div>
                        <div>F: 안전성 정책 위반 (점수 무관)</div>
                      </div>
                      <div class="pt-1 text-[10.5px] border-t border-border/60">※ 사실 오류 및 안전성 위반(⚠)은 경고 열에 표기됩니다.</div>
                    </div>
                  </div>
                </div>
              </div>
            </th>
            <th class="p-3 text-center whitespace-nowrap">경고</th>
            <th class="p-3 w-14 text-center">상세</th>
          </tr>
        </thead>
        <tbody class="divide-y">
          {#if loading && evaluations.length === 0}
            <tr>
              <td colspan="6" class="p-8 text-center text-muted-foreground">불러오는 중...</td>
            </tr>
          {:else if evaluations.length === 0}
            <tr>
              <td colspan="6" class="p-8 text-center text-muted-foreground">평가 내역이 없습니다.</td>
            </tr>
          {:else}
            {#each evaluations as item (item.id)}
              {@const isExpanded = expandedId === item.id}
              {@const judgment = isExpanded ? parseJSON<Judgment>(item.raw_json) : null}
              {@const rubrics = isExpanded ? parseJSON<Rubrics>(item.rubrics) : null}
              <tr class="hover:bg-muted/30 transition cursor-pointer" onclick={() => toggleExpand(item)}>
                <td class="p-3 text-xs text-muted-foreground whitespace-nowrap" title={formatDate(item.created_at)}
                  >{getAge(item.created_at, now)}</td
                >
                <td class="p-3">
                  <a
                    href={item.page_id > 0 ? `/w/index.php?curid=${item.page_id}` : getWikiViewHref(item.page_title)}
                    target="_blank"
                    rel="noopener"
                    class="font-medium hover:underline"
                    onclick={(e) => e.stopPropagation()}
                  >
                    {displayTitle(item.page_title)}
                  </a>
                </td>
                <td class="p-3 text-center whitespace-nowrap">
                  <span
                    class={`inline-flex items-center justify-center w-7 py-0.5 rounded text-sm font-bold border ${getGradeBadgeClass(item.grade)}`}
                  >
                    {item.grade}
                  </span>
                </td>
                <td class="p-3 text-center whitespace-nowrap">
                  {getScore(item)}
                </td>
                <td class="p-3 text-center whitespace-nowrap">
                  {#if item.safety_violated || item.accuracy_contradicted > 0}
                    <div class="inline-flex items-center justify-center gap-1.5">
                      {#if item.safety_violated}
                        <span
                          class="inline-flex items-center text-a-rose-600 font-medium text-xs"
                          title={item.safety_policies ? `안전성 정책 위반: ${item.safety_policies}` : '안전성 정책 위반'}
                        >
                          ⚠
                        </span>
                      {/if}
                      {#if item.accuracy_contradicted > 0}
                        <span
                          class="inline-flex items-center text-a-orange-600 font-medium text-xs"
                          title={`틀린 문장 ${item.accuracy_contradicted}건`}
                        >
                          ⚠ {item.accuracy_contradicted}
                        </span>
                      {/if}
                    </div>
                  {:else}
                    <span class="text-muted-foreground/60">-</span>
                  {/if}
                </td>
                <td class="p-3 text-center">
                  <button
                    type="button"
                    class="text-muted-foreground hover:text-foreground transition p-1 inline-flex items-center justify-center rounded hover:bg-muted"
                    aria-label={isExpanded ? '접기' : '펼치기'}
                    onclick={(e) => {
                      e.stopPropagation()
                      toggleExpand(item)
                    }}
                  >
                    <ZIcon path={mdiChevronDown} size={18} class={`transition-transform duration-200 ${isExpanded ? 'rotate-180' : ''}`} />
                  </button>
                </td>
              </tr>

              {#if isExpanded}
                {@const revContent = item.revid ? revisionContents[item.revid] : null}
                <tr class="bg-muted/20">
                  <td colspan="6" class="p-4 space-y-3">
                    <!-- 3분할: 1열 wikitext / 2열 총평·안전성·정확성·텍스트 품질 / 3열 모델·전반적 품질 (세로 실선, 외곽선 없음) -->
                    <div class="grid grid-cols-1 lg:grid-cols-3 divide-y lg:divide-y-0 lg:divide-x divide-border items-start">
                      <!-- 1열: wikitext -->
                      <div class="flex flex-col space-y-2 lg:pr-4 pb-4 lg:pb-0">
                        <div class="flex items-center gap-1 px-2.5 py-1 rounded bg-muted/70 text-xs text-muted-foreground font-medium">
                          {#if item.revid > 0}
                            <a
                              href={item.page_id > 0
                                ? `/w/index.php?curid=${item.page_id}&oldid=${item.revid}`
                                : `/w/index.php?oldid=${item.revid}`}
                              target="_blank"
                              rel="noopener"
                              class="hover:underline"
                              title="평가 당시 판 보기"
                            >
                              #{item.revid}판
                            </a>
                          {/if}
                          {#if item.is_latest}
                            <span>(최신)</span>
                          {:else if item.revisions_after !== undefined && item.revisions_after > 0}
                            <a
                              href={item.page_id > 0
                                ? `/w/index.php?curid=${item.page_id}&diff=cur&oldid=${item.revid}`
                                : `/w/index.php?oldid=${item.revid}&diff=cur`}
                              target="_blank"
                              rel="noopener"
                              class="hover:underline"
                              title="최신 판과 차이 비교"
                            >
                              (+{item.revisions_after}회 수정됨)
                            </a>
                          {/if}
                        </div>

                        <div
                          class="max-h-[560px] overflow-y-auto rounded border border-border/60 bg-muted/40 dark:bg-muted/15 p-2.5 font-mono text-xs leading-relaxed whitespace-pre-wrap break-all select-text"
                        >
                          {#if !item.revid || item.revid <= 0}
                            <div class="text-muted-foreground italic py-8 text-center">리비전 정보가 없습니다.</div>
                          {:else if revContent?.loading}
                            <div class="flex items-center justify-center py-16 text-muted-foreground">
                              <span class="animate-pulse">리비전 본문 불러오는 중...</span>
                            </div>
                          {:else if revContent?.error}
                            <div class="flex flex-col items-center justify-center py-12 text-a-rose-600 gap-2">
                              <span>{revContent.error}</span>
                              <button
                                type="button"
                                class="text-xs px-2.5 py-1 rounded border bg-background hover:bg-muted"
                                onclick={() => fetchRevisionContent(item.revid)}
                              >
                                다시 시도
                              </button>
                            </div>
                          {:else if revContent?.text}
                            {@const split = splitAtBytes(revContent.text, 6000)}
                            <span>{split.evaluated}</span>
                            {#if split.remaining}
                              <div
                                class="my-3 py-1.5 px-3 border-y border-dashed border-a-amber-300 bg-a-amber-100 text-[11px] text-a-amber-800 font-sans font-medium text-center"
                              >
                                ✂ 여기까지만 평가 모델에 전달됨 (앞 6,000 바이트)
                              </div>
                              <span class="text-muted-foreground/60">{split.remaining}</span>
                            {/if}
                          {:else}
                            <div class="text-muted-foreground italic py-8 text-center">본문 내용이 없습니다.</div>
                          {/if}
                        </div>
                      </div>

                      <!-- 2열: 대상 정의, 총평, 안전성, 정확성, 텍스트 품질 -->
                      <div class="space-y-3 text-xs py-4 lg:py-0 lg:px-4">
                        <!-- 대상 정의 -->
                        {#if item.definition}
                          <div class="space-y-1.5">
                            <div class="px-2.5 py-1 rounded bg-muted/70 font-semibold text-foreground">대상 정의</div>
                            <div class="text-muted-foreground pt-0.5 leading-relaxed select-text break-words">{item.definition}</div>
                          </div>
                        {/if}

                        <!-- 총평 -->
                        <div class="space-y-1.5">
                          <div class="px-2.5 py-1 rounded bg-muted/70 font-semibold text-foreground">
                            총평 (등급: {item.grade}, 점수: {getScore(item)})
                          </div>
                          <div class="text-muted-foreground pt-0.5 leading-relaxed select-text whitespace-pre-wrap break-words">
                            {item.summary}
                          </div>
                        </div>

                        <!-- 안전성 -->
                        {#if judgment?.safety || item.safety_violated}
                          {@const isViolated = item.safety_violated || judgment?.safety?.violated}
                          <div class={`space-y-1.5 ${isViolated ? 'text-a-rose-700' : ''}`}>
                            <div class="px-2.5 py-1 rounded bg-muted/70 font-semibold text-foreground">안전성</div>
                            <div class="text-muted-foreground pt-0.5">
                              {#if isViolated}
                                <span class="font-medium text-a-rose-600">
                                  위반 정책: {item.safety_policies
                                    .split(',')
                                    .filter(Boolean)
                                    .map((p) => policyLabels[p] ?? p)
                                    .join(', ')}
                                </span>
                                {#if judgment?.safety?.reason}
                                  <div class="text-muted-foreground pt-1">
                                    <span class="text-a-rose-600 mr-1">⚠</span>{judgment.safety.reason}
                                  </div>
                                {/if}
                              {:else}
                                <span>위반 정책 없음</span>
                                {#if judgment?.safety?.reason}
                                  <div class="text-muted-foreground/80 pt-0.5">{judgment.safety.reason}</div>
                                {/if}
                              {/if}
                            </div>
                          </div>
                        {/if}

                        <!-- 정확성 -->
                        {#if judgment?.accuracy}
                          <div class={`space-y-1.5 ${item.accuracy_contradicted > 0 ? 'text-a-orange-700' : ''}`}>
                            <div class="px-2.5 py-1 rounded bg-muted/70 font-semibold text-foreground">정확성</div>
                            <div class="text-muted-foreground pt-0.5">
                              맞음 {judgment.accuracy.supported ?? 0} · 틀림 {judgment.accuracy.contradictory ?? 0} · 사실 주장 없음 {judgment
                                .accuracy.no_rad ?? 0}
                            </div>
                            {#if judgment.accuracy.errors && judgment.accuracy.errors.length > 0}
                              <ul class="list-none pl-0 space-y-1 text-muted-foreground pt-1">
                                {#each judgment.accuracy.errors as e, i (i)}
                                  <li class="list-none">
                                    <span class="text-a-orange-600 mr-1">⚠</span><span class="text-foreground">"{e.sentence}"</span> — {e.reason}
                                  </li>
                                {/each}
                              </ul>
                            {/if}
                          </div>
                        {/if}

                        <!-- 텍스트 품질 -->
                        <div class="space-y-2">
                          <div class="px-2.5 py-1 rounded bg-muted/70 font-semibold text-foreground">
                            텍스트 품질 {item.text_quality}%
                          </div>
                          <ul class="list-none pl-0 space-y-1.5 pt-1">
                            {#each criteriaWithVerdicts(rubrics?.text_quality, judgment?.text_quality) as row (row.id)}
                              <li class="list-none">
                                <span class={row.pass ? 'text-a-emerald-700' : 'text-a-rose-700'}>{row.pass ? '✓' : '✗'}</span>
                                <span class="text-foreground">{row.criterion}</span>
                                {#if row.reason}
                                  <div class="text-muted-foreground pl-4">{row.reason}</div>
                                {/if}
                              </li>
                            {/each}
                          </ul>
                        </div>
                      </div>

                      <!-- 3열: 모델, 전반적 품질 -->
                      <div class="space-y-3 text-xs pt-4 lg:pt-0 lg:pl-4">
                        <!-- 모델 정보 -->
                        <div class="text-[11px] text-muted-foreground space-y-0.5">
                          <div>모델: <ModelId id={item.model} /></div>
                          <div class="flex flex-wrap items-center gap-x-3">
                            <span>입력: {item.prompt_tokens}t</span>
                            <span>출력: {item.completion_tokens}t</span>
                            <span>소요: {(item.elapsed_ms / 1000).toFixed(2)}초</span>
                          </div>
                        </div>

                        <!-- 전반적 품질 -->
                        <div class="space-y-2">
                          <div class="px-2.5 py-1 rounded bg-muted/70 font-semibold text-foreground">
                            전반적 품질 {item.general_quality}%
                          </div>
                          <ul class="list-none pl-0 space-y-1.5 pt-1">
                            {#each criteriaWithVerdicts(rubrics?.general_quality, judgment?.general_quality) as row (row.id)}
                              <li class="list-none">
                                <span class={row.pass ? 'text-a-emerald-700' : 'text-a-rose-700'}>{row.pass ? '✓' : '✗'}</span>
                                <span class="text-foreground">{row.criterion}</span>
                                {#if row.reason}
                                  <div class="text-muted-foreground pl-4">{row.reason}</div>
                                {/if}
                              </li>
                            {/each}
                          </ul>
                        </div>
                      </div>
                    </div>
                  </td>
                </tr>
              {/if}
            {/each}
          {/if}
        </tbody>
      </table>
    </div>

    <!-- 페이지네이션 -->
    {#if totalCount > pageSize}
      <div class="flex items-center justify-between p-3 border-t bg-muted/20 text-xs">
        <span class="text-muted-foreground">
          총 {totalCount}건 중 {pageOffset + 1} - {Math.min(pageOffset + pageSize, totalCount)}건
        </span>
        <div class="flex gap-2">
          <CButton
            size="small"
            variant="outline"
            disabled={pageOffset === 0}
            onclick={() => {
              pageOffset = Math.max(0, pageOffset - pageSize)
              fetchEvaluations()
            }}
          >
            이전
          </CButton>
          <CButton
            size="small"
            variant="outline"
            disabled={pageOffset + pageSize >= totalCount}
            onclick={() => {
              pageOffset += pageSize
              fetchEvaluations()
            }}
          >
            다음
          </CButton>
        </div>
      </div>
    {/if}
  </div>
</div>

<ZModal
  show={showRunModal}
  title="문서 평가 즉시 실행"
  okText="즉시 실행"
  okVariant="default"
  okDisabled={checkState !== 'exists' || triggeringRun || fetchingCandidate}
  panelClass="max-w-md !overflow-visible"
  sectionClass="p-5 !overflow-visible"
  onOk={handleConfirmRun}
  onCancel={() => {
    showRunModal = false
    resetRunModal()
  }}
>
  <div class="space-y-2 text-sm">
    <div class="flex items-center gap-2">
      <button
        type="button"
        class="shrink-0 h-9 rounded border bg-background px-3 text-xs font-medium text-foreground transition hover:bg-muted cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
        disabled={fetchingCandidate || triggeringRun}
        onclick={fetchAutoCandidate}
      >
        {#if fetchingCandidate}
          선정 중...
        {:else}
          자동선정
        {/if}
      </button>
      <div class="relative min-w-0 flex-1">
        <input
          type="text"
          bind:this={titleInputEl}
          value={targetTitle}
          class={`w-full rounded border bg-background p-2 pr-8 text-sm outline-none transition-colors ${
            checkState === 'exists'
              ? 'border-a-emerald-500 focus:border-a-emerald-600'
              : checkState === 'missing'
                ? 'border-a-rose-500 focus:border-a-rose-600'
                : checkState === 'checking'
                  ? 'border-a-blue-500 focus:border-a-blue-600'
                  : 'border-input focus:border-primary'
          }`}
          placeholder="또는 직접 문서 제목 입력"
          autocomplete="off"
          oninput={onTitleInput}
          onfocus={() => {
            if (suggestions.length > 0) showSuggestions = true
          }}
          onblur={() => {
            setTimeout(() => {
              showSuggestions = false
              activeSuggestionIndex = -1
            }, 200)
          }}
          onkeydown={onTitleKeydown}
        />
        <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2.5">
          {#if checkState === 'checking'}
            <span class="inline-block h-3.5 w-3.5 animate-spin rounded-full border-2 border-a-blue-500 border-t-transparent"></span>
          {:else if checkState === 'exists'}
            <span class="text-xs font-bold text-a-emerald-600" title="문서 확인됨">✓</span>
          {:else if checkState === 'missing'}
            <span class="text-xs font-bold text-a-rose-600" title="문서 없음">✗</span>
          {/if}
        </div>

        {#if showSuggestions && suggestions.length > 0}
          <div
            class="absolute left-0 right-0 top-full z-50 mt-1 max-h-56 overflow-y-auto rounded-md border bg-background shadow-lg text-xs"
          >
            {#each suggestions as item, i (item.id)}
              <button
                type="button"
                class={`flex w-full items-center justify-between px-3 py-2 text-left transition hover:bg-muted cursor-pointer ${
                  i === activeSuggestionIndex ? 'bg-muted text-primary font-medium' : 'text-foreground'
                }`}
                onmousedown={(e) => {
                  e.preventDefault()
                  selectSuggestion(item)
                }}
              >
                <div class="flex flex-col min-w-0">
                  <span class="truncate font-medium">{item.title}</span>
                  {#if item.matched_title}
                    <span class="text-[11px] text-muted-foreground truncate">{item.matched_title}에서 넘어옴</span>
                  {:else if item.description}
                    <span class="text-[11px] text-muted-foreground truncate">{item.description}</span>
                  {/if}
                </div>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
    {#if checkState === 'missing'}
      <p class="text-xs text-a-rose-500">
        '{trimmedTitle}' 문서를 찾을 수 없습니다.
      </p>
    {/if}
  </div>
</ZModal>
