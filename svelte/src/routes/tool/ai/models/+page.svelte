<svelte:options runes={true} />

<script lang="ts">
  import { onMount } from 'svelte'

  import { mdiFlash } from '@mdi/js'

  import ZIcon from '$shared/ui/ZIcon.svelte'
  import httpy from '$shared/utils/httpy'

  interface ModelInfo {
    id: string
    provider: string
    enabled: boolean
  }

  interface ModelUsage {
    id: string
    provider: string
    requests: number
    successes: number
    failures: number
    daily_limit: number | null
    remaining: number | null
    available: boolean
    cooldown_until?: string
    quota_source?: 'provider' | 'estimate'
    quota_reset?: string
    last_error?: string
  }

  // zengine's own record of a model over all its chain calls (llm_calls); unlike bob's usage it survives restarts.
  interface ModelHistory {
    model: string
    calls: number
    oks: number
    avg_elapsed_ms: number
    first_call_at: string | null
    last_ok_at: string | null
    last_failure_at: string | null
    last_error: string
    chains: string
  }

  interface ModelsResp {
    models: { data: ModelInfo[]; errors?: string[] }
    usage: { quota_day: string; next_reset: string; data: ModelUsage[] } | null
    history: ModelHistory[] | null
  }

  let loading = $state(true)
  let loadError = $state('')
  let resp = $state<ModelsResp | null>(null)

  interface Row {
    id: string
    provider: string
    model: string
    order?: number
    usage?: ModelUsage
  }

  function toRow(id: string, provider: string, order?: number, usage?: ModelUsage): Row {
    const slash = id.indexOf('/')
    return { id, provider, model: slash > 0 ? id.slice(slash + 1) : id, order, usage }
  }

  // Configured models (numbered in fallback order, the usage list's order) and any other enabled ones, most calls first.
  let configuredRows = $derived.by(() => {
    if (!resp) return []
    const usage = resp.usage?.data ?? []
    const rows = usage.map((u, i) => toRow(u.id, u.provider, i + 1, u))
    const seen = new Set(rows.map((r) => r.id))
    for (const m of resp.models.data ?? []) {
      if (m.enabled && !seen.has(m.id)) rows.push(toRow(m.id, m.provider))
    }
    return rows.sort((a, b) => (b.usage?.requests ?? 0) - (a.usage?.requests ?? 0))
  })

  // Models the providers offer that bob is not configured to use, by provider and name.
  let otherRows = $derived(
    (resp?.models.data ?? [])
      .filter((m) => !m.enabled)
      .map((m) => toRow(m.id, m.provider))
      .sort((a, b) => a.provider.localeCompare(b.provider) || a.model.localeCompare(b.model)),
  )

  // yyyy-MM-dd HH:mm rounded to the nearest minute (a reset at 15:59:59.9 shows as 16:00); HH:mm alone for today.
  function formatTime(iso?: string): string {
    if (!iso) return ''
    const t = new Date(iso).getTime()
    if (isNaN(t)) return iso
    const d = new Date(Math.round(t / 60000) * 60000)
    const pad = (n: number) => String(n).padStart(2, '0')
    const time = `${pad(d.getHours())}:${pad(d.getMinutes())}`
    if (d.toDateString() === new Date().toDateString()) return time
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${time}`
  }

  const kstFormat = new Intl.DateTimeFormat('en-US', {
    timeZone: 'Asia/Seoul',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hourCycle: 'h23',
  })

  // Start of the Pacific-time quota day (yyyy-MM-dd) as a timestamp: 00:00 PST is 08:00 UTC, 00:00 PDT is 07:00 UTC.
  function pacificMidnight(day: string): number {
    const [y, m, d] = day.split('-').map(Number)
    const t = Date.UTC(y, m - 1, d, 8)
    const hour = new Intl.DateTimeFormat('en-US', { timeZone: 'America/Los_Angeles', hour: 'numeric', hourCycle: 'h23' }).format(t)
    return Number(hour) === 0 ? t : t - 3600_000
  }

  // MM.dd and HH:mm in KST, rounded to the nearest minute.
  function kstParts(t: number | string): { date: string; time: string } {
    const d = new Date(t)
    if (isNaN(d.getTime())) return { date: '', time: '' }
    const parts = kstFormat.formatToParts(new Date(Math.round(d.getTime() / 60000) * 60000))
    const get = (type: string) => parts.find((p) => p.type === type)?.value ?? ''
    return { date: `${get('month')}.${get('day')}`, time: `${get('hour')}:${get('minute')}` }
  }

  // The quota day in KST, such as "10.23 - 10.24 16:00"; the start time is shown only when it differs (a DST change day).
  function formatQuotaPeriod(day: string, nextReset: string): string {
    const start = kstParts(pacificMidnight(day))
    const end = kstParts(nextReset)
    const from = start.time === end.time ? start.date : `${start.date} ${start.time}`
    return `${from} - ${end.date} ${end.time}`
  }

  async function fetchModels() {
    loading = true
    const [data, err] = await httpy.get<ModelsResp>('/api/ai/models')
    loading = false
    if (err) {
      loadError = err.message || '모델 정보를 불러오지 못했습니다.'
      return
    }
    loadError = ''
    resp = data
  }

  onMount(() => {
    fetchModels()
  })
</script>

{#snippet modelRow(row: Row, n: number)}
  {@const u = row.usage}
  <tr class={row.order ? '' : 'text-muted-foreground'}>
    <td class="p-2 text-center text-muted-foreground">{n}</td>
    <td class="p-2">{row.provider}</td>
    <td class="p-2">{row.model}</td>
    <td class="p-2 text-xs">
      {#if !u}
        {#if row.order}<span class="text-muted-foreground">-</span>{:else}<span aria-label="불가">●</span>{/if}
      {:else if u.available}
        <span class="text-emerald-600 dark:text-emerald-400" aria-label="가능">●</span>
      {:else}
        <span class="text-rose-600 dark:text-rose-400"
          >소진 · <ZIcon path={mdiFlash} size={14} class="inline-block align-text-bottom" />{formatTime(u.cooldown_until)}</span
        >
      {/if}
    </td>
    <td class="p-2 text-right">
      {#if u?.requests}{u.successes} <span class="text-xs text-muted-foreground">({u.requests})</span>{:else}-{/if}
    </td>
    <td class="p-2 text-right">{u ? (u.daily_limit ?? '-') : '-'}</td>
    <td class="p-2 text-right">
      {#if u}
        {u.remaining ?? '-'}
        {#if u.quota_source}
          <span class="text-xs text-muted-foreground">({u.quota_source})</span>
        {/if}
      {:else}
        -
      {/if}
    </td>
  </tr>
{/snippet}

<div class="container mx-auto p-4 md:p-6 max-w-6xl space-y-4">
  {#if loading && !resp}
    <div class="p-8 text-center text-muted-foreground">불러오는 중...</div>
  {:else if loadError}
    <div class="p-8 text-center text-rose-600 dark:text-rose-400">{loadError}</div>
  {:else if resp}
    <div class={`text-xs text-muted-foreground ${resp.usage ? 'text-right' : ''}`}>
      {#if resp.usage}
        {formatQuotaPeriod(resp.usage.quota_day, resp.usage.next_reset)}
      {:else}
        이 bob에는 사용량 조회(/aigate/v1/usage)가 없습니다.
      {/if}
    </div>
    {#if resp.models.errors?.length}
      <div class="text-xs text-rose-600 dark:text-rose-400">{resp.models.errors.join(' / ')}</div>
    {/if}

    <div class="rounded-lg border bg-card shadow-sm overflow-x-auto">
      <table class="w-full text-left text-sm border-collapse">
        <thead>
          <tr class="border-b bg-muted/50 text-muted-foreground text-xs">
            <th class="p-2 w-12 text-center">#</th>
            <th class="p-2">Provider</th>
            <th class="p-2">Model</th>
            <th class="p-2">Status</th>
            <th class="p-2 text-right">Calls</th>
            <th class="p-2 text-right">Limit</th>
            <th class="p-2 text-right">Remaining</th>
          </tr>
        </thead>
        <tbody class="divide-y">
          {#each configuredRows as row, i (row.id)}
            {@render modelRow(row, i + 1)}
          {/each}
          {#each otherRows as row, i (row.id)}
            {@render modelRow(row, configuredRows.length + i + 1)}
          {/each}
        </tbody>
      </table>
    </div>

    <h2 class="pt-2 text-sm font-semibold">
      호출 이력 <span class="font-normal text-xs text-muted-foreground">zengine 체인 호출 전체 (llm_calls)</span>
    </h2>
    {#if resp.history?.length}
      <div class="rounded-lg border bg-card shadow-sm overflow-x-auto">
        <table class="w-full text-left text-sm border-collapse">
          <thead>
            <tr class="border-b bg-muted/50 text-muted-foreground text-xs">
              <th class="p-2">Model</th>
              <th class="p-2 text-right">Calls</th>
              <th class="p-2 text-right">Success</th>
              <th class="p-2 text-right">Avg</th>
              <th class="p-2">Last OK</th>
              <th class="p-2">Last failure</th>
              <th class="p-2">Chains</th>
            </tr>
          </thead>
          <tbody class="divide-y">
            {#each resp.history as h (h.model)}
              <tr>
                <td class="p-2">{h.model}</td>
                <td class="p-2 text-right">{h.calls}</td>
                <td class="p-2 text-right">
                  {Math.round((h.oks / h.calls) * 100)}% <span class="text-xs text-muted-foreground">({h.oks})</span>
                </td>
                <td class="p-2 text-right">{(h.avg_elapsed_ms / 1000).toFixed(1)}s</td>
                <td class="p-2 text-xs">{h.last_ok_at ? formatTime(h.last_ok_at) : '-'}</td>
                <td class="p-2 text-xs">
                  {#if h.last_failure_at}
                    <div>{formatTime(h.last_failure_at)}</div>
                    {#if h.last_error}<div class="text-rose-600 dark:text-rose-400 break-words max-w-md">{h.last_error}</div>{/if}
                  {:else}
                    -
                  {/if}
                </td>
                <td class="p-2 text-xs text-muted-foreground">{h.chains}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <div class="text-xs text-muted-foreground">기록된 호출이 없습니다.</div>
    {/if}
  {/if}
</div>
