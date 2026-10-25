<svelte:options runes={true} />

<script lang="ts">
  import { onMount } from 'svelte'

  import { page } from '$app/state'
  import CButton from '$shared/ui/CButton.svelte'
  import { showToast } from '$shared/ui/toast/toast'
  import httpy from '$shared/utils/httpy'
  import { displayTitle } from '$shared/utils/mediawiki'
  import { getAge } from '$shared/utils/time'
  import { getWikiViewHref } from '$shared/utils/wikiLink'

  import LoopPanel from '../LoopPanel.svelte'
  import { fetchRevisionSource, formatDateTime, type LastRun, type LoopSettings, type QueuedTask, splitAtBytes } from '../loop'
  import ModelId from '../ModelId.svelte'
  import RunModal from '../RunModal.svelte'

  // The define step reads this much of the wikitext (services/doceval maxLeadLen).
  const leadBytes = 1500

  interface Definition {
    page_id: number
    page_title: string
    revid: number
    definition: string
    model: string
    chain_version: string
    updated_at: string
  }

  interface ListResp {
    total: number
    items: Definition[]
  }

  interface SettingsResp {
    define_interval: string
    define_model: string
    define_strategy: string
    define_last_run?: LastRun | null
    define_tasks?: QueuedTask[] | null
  }

  type SavedResp = SettingsResp & { ok: boolean }

  const pageSize = 20

  let settings = $state<LoopSettings>({ interval: '30m', model: 'groq/openai/gpt-oss-120b', strategy: 'random' })
  let savingSettings = $state(false)
  let settingsError = $state('')
  let lastRun = $state<LastRun | null>(null)
  let tasks = $state<QueuedTask[]>([])
  let now = $state(Date.now())

  let items = $state<Definition[]>([])
  let total = $state(0)
  let loading = $state(true)
  let offset = $state(0)
  let query = $state(page.url.searchParams.get('q')?.replaceAll('_', ' ') ?? '')
  let expandedId = $state<number | null>(null)
  let sources = $state<Record<number, { loading: boolean; text?: string; error?: string }>>({})

  let showRunModal = $state(false)
  let triggeringRun = $state(false)

  function applySettings(data: SettingsResp) {
    settings = { interval: data.define_interval, model: data.define_model, strategy: data.define_strategy }
  }

  // After the first load, a newly seen successful run (also the first one ever) refreshes the list.
  let settingsLoaded = false

  async function fetchSettings() {
    const [data, err] = await httpy.get<SettingsResp>('/api/evaluations/settings')
    if (err || !data) return
    applySettings(data)
    tasks = data.define_tasks ?? []
    const prevAt = lastRun?.at
    lastRun = data.define_last_run ?? null
    if (settingsLoaded && lastRun?.ok && lastRun.at !== prevAt) fetchDefinitions()
    settingsLoaded = true
  }

  async function handleSaveSettings(change: Partial<Omit<LoopSettings, 'model'>>) {
    savingSettings = true
    settingsError = ''
    const body = Object.fromEntries(Object.entries(change).map(([k, v]) => [`define_${k}`, v]))
    const [data, err] = await httpy.put<SavedResp>('/api/evaluations/settings', body)
    savingSettings = false
    if (err) {
      settingsError = err.message || '설정 저장에 실패했습니다.'
      return
    }
    if (data?.ok) {
      applySettings(data)
      showToast('저장했습니다')
    }
  }

  async function runDefinition(pageID: number, title: string) {
    triggeringRun = true
    const [, err] = await httpy.post<{ ok: boolean; task_id: string }>('/api/definitions/run', { page_id: pageID, page_title: title })
    triggeringRun = false
    if (err) {
      showToast(err.message || '정의 실행에 실패했습니다.')
      return
    }
    showToast(`'${title}' 정의를 시작했습니다`)
    fetchSettings()
  }

  async function fetchDefinitions() {
    loading = true
    const params: Record<string, string> = { limit: String(pageSize), offset: String(offset) }
    if (query.trim()) params.q = query.trim()
    const [data, err] = await httpy.get<ListResp>('/api/definitions', params)
    loading = false
    if (err || !data) {
      items = []
      total = 0
      return
    }
    items = data.items ?? []
    total = data.total ?? 0
  }

  let searchTimer: ReturnType<typeof setTimeout> | undefined
  function onSearchInput(e: Event) {
    query = (e.target as HTMLInputElement).value
    clearTimeout(searchTimer)
    searchTimer = setTimeout(() => {
      offset = 0
      fetchDefinitions()
    }, 300)
  }

  async function toggleExpand(item: Definition) {
    if (expandedId === item.page_id) {
      expandedId = null
      return
    }
    expandedId = item.page_id
    if (item.revid <= 0 || sources[item.revid]) return
    sources[item.revid] = { loading: true }
    const src = await fetchRevisionSource(item.revid)
    sources[item.revid] = typeof src === 'string' ? { loading: false, text: src } : { loading: false, error: src.message }
  }

  onMount(() => {
    fetchSettings()
    fetchDefinitions()
    let tick = 0
    const timer = setInterval(() => {
      now = Date.now()
      if (++tick % 5 === 0) fetchSettings()
    }, 1000)
    return () => {
      clearInterval(timer)
      clearTimeout(searchTimer)
    }
  })
</script>

<div class="container mx-auto p-4 md:p-6 max-w-6xl space-y-6">
  <LoopPanel
    chain="doc-define"
    heading="문서 정의"
    instantTitle="정의 즉시 실행 대화상자 열기"
    {settings}
    saving={savingSettings}
    {lastRun}
    {tasks}
    {now}
    instantDisabled={triggeringRun}
    onchange={handleSaveSettings}
    oninstant={() => (showRunModal = true)}
  />

  {#if settingsError}
    <div class="text-xs font-medium text-a-rose-600">⚠️ {settingsError}</div>
  {/if}

  <div class="flex flex-wrap items-center gap-3 text-xs">
    <input
      type="search"
      class="h-8 w-full sm:w-64 rounded border bg-background px-2.5 focus:outline-none focus:ring-1 focus:ring-primary"
      placeholder="제목 검색"
      value={query}
      oninput={onSearchInput}
    />
    <span class="ml-auto text-muted-foreground">{total.toLocaleString()}건</span>
  </div>

  <div class="rounded-lg border bg-card overflow-hidden shadow-sm">
    <div class="overflow-x-auto">
      <table class="w-full text-left border-collapse text-sm">
        <thead>
          <tr class="border-b bg-muted/50 text-muted-foreground text-xs font-medium">
            <th class="p-3 whitespace-nowrap">일시</th>
            <th class="p-3">제목</th>
            <th class="p-3">정의</th>
            <th class="p-3 whitespace-nowrap">모델</th>
          </tr>
        </thead>
        <tbody class="divide-y">
          {#if loading && items.length === 0}
            <tr><td colspan="4" class="p-8 text-center text-muted-foreground">불러오는 중...</td></tr>
          {:else if items.length === 0}
            <tr><td colspan="4" class="p-8 text-center text-muted-foreground">정의가 없습니다.</td></tr>
          {:else}
            {#each items as item (item.page_id)}
              <tr class="hover:bg-muted/30 transition cursor-pointer align-top" onclick={() => toggleExpand(item)}>
                <td class="p-3 text-xs text-muted-foreground whitespace-nowrap" title={formatDateTime(item.updated_at)}
                  >{getAge(item.updated_at, now)}</td
                >
                <td class="p-3 whitespace-nowrap">
                  <a
                    href={item.page_id > 0 ? `/w/index.php?curid=${item.page_id}` : getWikiViewHref(item.page_title)}
                    target="_blank"
                    rel="noopener"
                    class="font-medium hover:underline"
                    onclick={(e) => e.stopPropagation()}>{displayTitle(item.page_title)}</a
                  >
                </td>
                <td class="p-3 break-words">{item.definition}</td>
                <td class="p-3 text-xs whitespace-nowrap"><ModelId id={item.model} /></td>
              </tr>
              {#if expandedId === item.page_id}
                {@const src = sources[item.revid]}
                <tr class="bg-muted/10">
                  <td colspan="4" class="p-4 text-xs">
                    <div class="mb-2 text-muted-foreground">
                      r{item.revid} · {item.chain_version} · 정의 단계가 읽은 앞부분 {leadBytes}바이트
                    </div>
                    {#if !src || src.loading}
                      <div class="text-muted-foreground">불러오는 중...</div>
                    {:else if src.error}
                      <div class="text-a-rose-500">{src.error}</div>
                    {:else if src.text !== undefined}
                      {@const split = splitAtBytes(src.text, leadBytes)}
                      <pre class="whitespace-pre-wrap break-words font-mono text-[11px] leading-relaxed max-h-96 overflow-y-auto"><span
                          >{split.read}</span
                        >{#if split.remaining}<span class="text-muted-foreground/50">{split.remaining}</span>{/if}</pre>
                    {/if}
                  </td>
                </tr>
              {/if}
            {/each}
          {/if}
        </tbody>
      </table>
    </div>

    {#if total > pageSize}
      <div class="flex items-center justify-between p-3 border-t bg-muted/20 text-xs">
        <span class="text-muted-foreground">총 {total}건 중 {offset + 1} - {Math.min(offset + pageSize, total)}건</span>
        <div class="flex gap-2">
          <CButton
            size="small"
            variant="outline"
            disabled={offset === 0}
            onclick={() => {
              offset = Math.max(0, offset - pageSize)
              fetchDefinitions()
            }}>이전</CButton
          >
          <CButton
            size="small"
            variant="outline"
            disabled={offset + pageSize >= total}
            onclick={() => {
              offset += pageSize
              fetchDefinitions()
            }}>다음</CButton
          >
        </div>
      </div>
    {/if}
  </div>
</div>

<RunModal
  bind:show={showRunModal}
  heading="문서 정의 즉시 실행"
  candidateUrl={`/api/definitions/candidate?strategy=${encodeURIComponent(settings.strategy)}`}
  busy={triggeringRun}
  onrun={runDefinition}
/>
