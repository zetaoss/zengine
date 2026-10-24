<svelte:options runes={true} />

<script lang="ts">
  import { onMount } from 'svelte'

  import { resolve } from '$app/paths'

  import { mdiChevronDown, mdiInformationOutline } from '@mdi/js'

  import CButton from '$shared/ui/CButton.svelte'
  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import httpy from '$shared/utils/httpy'
  import { displayTitle } from '$shared/utils/mediawiki'
  import { getAge } from '$shared/utils/time'
  import { getWikiViewHref } from '$shared/utils/wikiLink'

  import LoopPanel from '../LoopPanel.svelte'
  import { fetchModelIds, fetchRevisionSource, type LastRun, type LoopSettings, type QueuedTask, splitAtBytes } from '../loop'
  import ModelId from '../ModelId.svelte'
  import RunModal from '../RunModal.svelte'

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

  interface SettingsResp extends LoopSettings {
    last_run?: LastRun | null
    tasks?: QueuedTask[] | null
  }

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

  let settings = $state<LoopSettings>({ interval: '1h', model: 'groq/openai/gpt-oss-120b', strategy: 'random' })
  let modelIds = $state<string[]>([])
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

  let queuedTasks = $state<QueuedTask[]>([])
  let lastRun = $state<LastRun | null>(null)
  let now = $state(Date.now())
  let initialSettingsLoaded = false

  async function fetchSettings() {
    const [data, err] = await httpy.get<SettingsResp>('/api/evaluations/settings')
    if (!err && data) {
      settings = { interval: data.interval, model: data.model, strategy: data.strategy }
      queuedTasks = data.tasks ?? []
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

  async function handleSaveSettings(change: Partial<LoopSettings>) {
    savingSettings = true
    settingsError = ''

    const [data, err] = await httpy.put<LoopSettings & { ok: boolean }>('/api/evaluations/settings', change)
    savingSettings = false

    if (err) {
      settingsError = err.message || '설정 저장에 실패했습니다.'
      return
    }

    if (data?.ok) {
      settings = { interval: data.interval, model: data.model, strategy: data.strategy }
      showToast('저장했습니다')
    }
  }

  let showRunModal = $state(false)
  let triggeringRun = $state(false)

  async function runEvaluation(pageID: number, title: string) {
    triggeringRun = true
    const [, err] = await httpy.post<{ ok: boolean; task_id: string }>('/api/evaluations/run', { page_id: pageID, page_title: title })
    triggeringRun = false
    if (err) {
      showToast(err.message || '평가 실행에 실패했습니다.')
      return
    }
    showToast(`'${title}' 평가를 시작했습니다`)
    fetchSettings()
  }

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

  async function fetchRevisionContent(revid: number) {
    if (!revid || revid <= 0) return
    if (revisionContents[revid]?.text || revisionContents[revid]?.loading) return
    revisionContents[revid] = { loading: true }
    const src = await fetchRevisionSource(revid)
    revisionContents[revid] = typeof src === 'string' ? { loading: false, text: src } : { loading: false, error: src.message }
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
    fetchModelIds().then((ids) => (modelIds = ids))
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
  <LoopPanel
    heading="문서 평가"
    instantTitle="평가 즉시 실행 대화상자 열기"
    {settings}
    {modelIds}
    saving={savingSettings}
    {lastRun}
    tasks={queuedTasks}
    {now}
    instantDisabled={triggeringRun}
    onchange={handleSaveSettings}
    oninstant={() => (showRunModal = true)}
  />

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
                            <span>{split.read}</span>
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
                            <div class="px-2.5 py-1 rounded bg-muted/70 font-semibold text-foreground flex items-center justify-between">
                              <span>대상 정의</span>
                              <a
                                href={`${resolve('/tool/ai/define')}?q=${encodeURIComponent(item.page_title)}`}
                                class="font-normal text-muted-foreground hover:underline"
                                onclick={(e) => e.stopPropagation()}>정의 목록</a
                              >
                            </div>
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

<RunModal
  bind:show={showRunModal}
  heading="문서 평가 즉시 실행"
  candidateUrl={`/api/evaluations/candidate?strategy=${encodeURIComponent(settings.strategy)}`}
  busy={triggeringRun}
  onrun={runEvaluation}
/>
