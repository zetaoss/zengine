<svelte:options runes={true} />

<script lang="ts">
  import { onMount } from 'svelte'

  import ZButton from '$shared/ui/ZButton.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import httpy from '$shared/utils/httpy'

  import LineChart from './LineChart.svelte'

  type MetricKey = 'uniq_uniques' | 'sum_requests' | 'sum_bytes' | 'sum_cachedBytes'
  type ChartUnit = 'count' | 'bytes' | 'percent'

  interface AnalyticsResp {
    timeslots: string[]
    uniq_uniques: Array<unknown>
    sum_requests: Array<unknown>
    sum_bytes: Array<unknown>
    sum_cachedBytes: Array<unknown>
  }

  type MwMetricKey = 'pages' | 'articles' | 'edits' | 'images' | 'users' | 'activeusers' | 'admins' | 'jobs'
  interface MwStatisticsResp {
    timeslots: string[]
    pages: Array<unknown>
    articles: Array<unknown>
    edits: Array<unknown>
    images: Array<unknown>
    users: Array<unknown>
    activeusers: Array<unknown>
    admins: Array<unknown>
    jobs: Array<unknown>
  }

  interface RowDef {
    key: string
    label: string
    statText: string
    unit: ChartUnit
    series: Array<{ label: string; color: string; values: Array<number | null> }>
  }

  const EMPTY: AnalyticsResp = {
    timeslots: [],
    uniq_uniques: [],
    sum_requests: [],
    sum_bytes: [],
    sum_cachedBytes: [],
  }
  const EMPTY_MW: MwStatisticsResp = {
    timeslots: [],
    pages: [],
    articles: [],
    edits: [],
    images: [],
    users: [],
    activeusers: [],
    admins: [],
    jobs: [],
  }

  let loading = $state(true)
  let failed = $state<string | null>(null)
  let range = $state<'24h' | '7d' | '30d'>('24h')
  let syncedHoverIndex = $state<number | null>(null)
  let data = $state<AnalyticsResp>(EMPTY)
  let mwData = $state<MwStatisticsResp>(EMPTY_MW)
  let fetchVersion = 0

  const labels = $derived.by(() => (range === '24h' ? data.timeslots : data.timeslots.map((v) => normalizeDateKey(v))))
  const rows = $derived.by<RowDef[]>(() => buildRows(data))
  const labelsMw = $derived.by(() => mwData.timeslots.map((v) => normalizeDateKey(v)))
  const mwRows = $derived.by<RowDef[]>(() => buildMwRows(mwData))

  async function fetchData() {
    const version = ++fetchVersion
    loading = true
    failed = null

    if (range === '24h') {
      const [resp, err] = await httpy.get<AnalyticsResp>('/api/dash/cf-analytics/hourly')
      if (version !== fetchVersion) return
      if (err) {
        failed = err.message
        loading = false
        return
      }
      data = normalizeResp(resp)
      mwData = EMPTY_MW
      loading = false
      return
    }

    const days = range === '7d' ? 7 : 30
    const [[cfResp, cfErr], [mwResp, mwErr]] = await Promise.all([
      httpy.get<AnalyticsResp>(`/api/dash/cf-analytics/daily/${days}`),
      httpy.get<MwStatisticsResp>(`/api/dash/mw-statistics/daily/${days}`),
    ])
    if (version !== fetchVersion) return

    if (cfErr) {
      failed = cfErr.message
      loading = false
      return
    }
    if (mwErr) {
      failed = mwErr.message
      loading = false
      return
    }

    data = normalizeResp(cfResp)
    mwData = normalizeMwResp(mwResp)
    loading = false
  }

  function normalizeResp(input: AnalyticsResp | null): AnalyticsResp {
    if (!input) return EMPTY
    return {
      timeslots: Array.isArray(input.timeslots) ? input.timeslots.map((v) => String(v)) : [],
      uniq_uniques: Array.isArray(input.uniq_uniques) ? input.uniq_uniques : [],
      sum_requests: Array.isArray(input.sum_requests) ? input.sum_requests : [],
      sum_bytes: Array.isArray(input.sum_bytes) ? input.sum_bytes : [],
      sum_cachedBytes: Array.isArray(input.sum_cachedBytes) ? input.sum_cachedBytes : [],
    }
  }

  function normalizeMwResp(input: MwStatisticsResp | null): MwStatisticsResp {
    if (!input) return EMPTY_MW
    return {
      timeslots: Array.isArray(input.timeslots) ? input.timeslots.map((v) => String(v)) : [],
      pages: Array.isArray(input.pages) ? input.pages : [],
      articles: Array.isArray(input.articles) ? input.articles : [],
      edits: Array.isArray(input.edits) ? input.edits : [],
      images: Array.isArray(input.images) ? input.images : [],
      users: Array.isArray(input.users) ? input.users : [],
      activeusers: Array.isArray(input.activeusers) ? input.activeusers : [],
      admins: Array.isArray(input.admins) ? input.admins : [],
      jobs: Array.isArray(input.jobs) ? input.jobs : [],
    }
  }

  function buildRows(resp: AnalyticsResp): RowDef[] {
    return [
      {
        key: 'unique-visitors',
        label: 'Unique Visitors',
        statText: fmtCount(sumMetric(resp, 'uniq_uniques')),
        unit: 'count',
        series: [{ label: 'Unique Visitors', color: '#0891b2', values: seriesOf(resp, 'uniq_uniques') }],
      },
      {
        key: 'total-requests',
        label: 'Total Requests',
        statText: fmtCount(sumMetric(resp, 'sum_requests')),
        unit: 'count',
        series: [{ label: 'Total Requests', color: '#0891b2', values: seriesOf(resp, 'sum_requests') }],
      },
      {
        key: 'percent-cached',
        label: 'Percent Cached',
        statText: fmtPercent(percentCachedTotal(resp)),
        unit: 'percent',
        series: [{ label: 'Percent Cached', color: '#0891b2', values: percentCachedSeries(resp) }],
      },
      {
        key: 'data-served',
        label: 'Total Data Served',
        statText: fmtBytes(sumMetric(resp, 'sum_bytes')),
        unit: 'bytes',
        series: [{ label: 'Total Data Served', color: '#0891b2', values: seriesOf(resp, 'sum_bytes') }],
      },
      {
        key: 'data-cached',
        label: 'Data Cached',
        statText: fmtBytes(sumMetric(resp, 'sum_cachedBytes')),
        unit: 'bytes',
        series: [{ label: 'Data Cached', color: '#0891b2', values: seriesOf(resp, 'sum_cachedBytes') }],
      },
    ]
  }

  function buildMwRows(resp: MwStatisticsResp): RowDef[] {
    return [
      {
        key: 'mw-pages',
        label: 'Pages',
        statText: fmtCount(lastMetric(resp, 'pages')),
        unit: 'count',
        series: [{ label: 'Pages', color: '#0891b2', values: seriesOfMw(resp, 'pages') }],
      },
      {
        key: 'mw-articles',
        label: 'Articles',
        statText: fmtCount(lastMetric(resp, 'articles')),
        unit: 'count',
        series: [{ label: 'Articles', color: '#0891b2', values: seriesOfMw(resp, 'articles') }],
      },
      {
        key: 'mw-edits',
        label: 'Edits',
        statText: fmtCount(lastMetric(resp, 'edits')),
        unit: 'count',
        series: [{ label: 'Edits', color: '#0891b2', values: seriesOfMw(resp, 'edits') }],
      },
      {
        key: 'mw-images',
        label: 'Images',
        statText: fmtCount(lastMetric(resp, 'images')),
        unit: 'count',
        series: [{ label: 'Images', color: '#0891b2', values: seriesOfMw(resp, 'images') }],
      },
      {
        key: 'mw-users',
        label: 'Users',
        statText: fmtCount(lastMetric(resp, 'users')),
        unit: 'count',
        series: [{ label: 'Users', color: '#0891b2', values: seriesOfMw(resp, 'users') }],
      },
      {
        key: 'mw-activeusers',
        label: 'Active Users',
        statText: fmtCount(lastMetric(resp, 'activeusers')),
        unit: 'count',
        series: [{ label: 'Active Users', color: '#0891b2', values: seriesOfMw(resp, 'activeusers') }],
      },
      {
        key: 'mw-admins',
        label: 'Admins',
        statText: fmtCount(lastMetric(resp, 'admins')),
        unit: 'count',
        series: [{ label: 'Admins', color: '#0891b2', values: seriesOfMw(resp, 'admins') }],
      },
      {
        key: 'mw-jobs',
        label: 'Jobs',
        statText: fmtCount(lastMetric(resp, 'jobs')),
        unit: 'count',
        series: [{ label: 'Jobs', color: '#0891b2', values: seriesOfMw(resp, 'jobs') }],
      },
    ]
  }

  function seriesOf(resp: AnalyticsResp, key: MetricKey): Array<number | null> {
    const src = resp[key] ?? []
    return resp.timeslots.map((_, idx) => toNumber(src[idx] ?? null))
  }

  function sumMetric(resp: AnalyticsResp, key: MetricKey): number {
    let total = 0
    const src = resp[key] ?? []
    for (let i = 0; i < src.length; i += 1) {
      const value = toNumber(src[i] ?? null)
      if (value != null) total += value
    }
    return total
  }

  function seriesOfMw(resp: MwStatisticsResp, key: MwMetricKey): Array<number | null> {
    const src = resp[key] ?? []
    return resp.timeslots.map((_, idx) => toNumber(src[idx] ?? null))
  }

  function lastMetric(resp: MwStatisticsResp, key: MwMetricKey): number | null {
    const src = resp[key] ?? []
    for (let i = src.length - 1; i >= 0; i -= 1) {
      const value = toNumber(src[i] ?? null)
      if (value != null) return value
    }
    return null
  }

  function percentCachedSeries(resp: AnalyticsResp): Array<number | null> {
    return resp.timeslots.map((_, idx) => {
      const served = toNumber(resp.sum_bytes[idx] ?? null)
      const cached = toNumber(resp.sum_cachedBytes[idx] ?? null)
      if (served == null || served <= 0 || cached == null) return null
      return (cached / served) * 100
    })
  }

  function percentCachedTotal(resp: AnalyticsResp): number | null {
    const served = sumMetric(resp, 'sum_bytes')
    const cached = sumMetric(resp, 'sum_cachedBytes')
    if (!Number.isFinite(served) || served <= 0) return null
    return (cached / served) * 100
  }

  function toNumber(value: unknown): number | null {
    return typeof value === 'number' && Number.isFinite(value) ? value : null
  }

  function md(value: string | null | undefined) {
    if (!value) return '-'
    const dayPart = normalizeDateKey(value)
    const compact = dayPart.replace(/[^0-9]/g, '')
    if (compact.length >= 8) {
      const m = Number(compact.slice(4, 6))
      const d = Number(compact.slice(6, 8))
      if (!Number.isFinite(m) || !Number.isFinite(d)) return '-'
      return `${m}.${d}`
    }
    return '-'
  }

  function fmtCount(value: number | null | undefined) {
    if (typeof value !== 'number' || !Number.isFinite(value)) return '-'
    const abs = Math.abs(value)
    if (abs >= 1_000_000_000) return `${stripZero((value / 1_000_000_000).toFixed(1))}B`
    if (abs >= 1_000_000) return `${stripZero((value / 1_000_000).toFixed(1))}M`
    if (abs >= 1_000) return `${stripZero((value / 1_000).toFixed(1))}k`
    return `${Math.round(value).toLocaleString('en-US')}`
  }

  function fmtBytes(value: number | null | undefined) {
    if (typeof value !== 'number' || !Number.isFinite(value)) return '-'
    const units = ['B', 'Ki', 'Mi', 'Gi', 'Ti']
    let n = value
    let u = 0
    while (n >= 1024 && u < units.length - 1) {
      n /= 1024
      u += 1
    }
    return `${stripZero(n.toFixed(1))}${units[u]}`
  }

  function fmtPercent(value: number | null | undefined) {
    if (typeof value !== 'number' || !Number.isFinite(value)) return '-'
    return `${value.toFixed(1)}%`
  }

  function stripZero(value: string) {
    return value.endsWith('.0') ? value.slice(0, -2) : value
  }

  function normalizeDateKey(value: string) {
    return value.trim().slice(0, 10)
  }

  onMount(() => {
    void fetchData()
  })
</script>

<div class="p-5">
  <h2 class="m-0 text-2xl font-bold">대시보드</h2>

  <div class="mb-4 mt-3 flex flex-wrap items-center justify-between gap-3">
    <div class="flex gap-2">
      <ZButton
        size="medium"
        cooldown={0}
        color={range === '24h' ? 'primary' : 'default'}
        class={`px-3 py-1 ${range === '24h' ? 'font-semibold' : ''}`}
        onclick={() => {
          if (range !== '24h') {
            range = '24h'
            void fetchData()
          }
        }}
      >
        24 Hours
      </ZButton>
      <ZButton
        size="medium"
        cooldown={0}
        color={range === '7d' ? 'primary' : 'default'}
        class={`px-3 py-1 ${range === '7d' ? 'font-semibold' : ''}`}
        onclick={() => {
          if (range !== '7d') {
            range = '7d'
            void fetchData()
          }
        }}
      >
        7 Days
      </ZButton>
      <ZButton
        size="medium"
        cooldown={0}
        color={range === '30d' ? 'primary' : 'default'}
        class={`px-3 py-1 ${range === '30d' ? 'font-semibold' : ''}`}
        onclick={() => {
          if (range !== '30d') {
            range = '30d'
            void fetchData()
          }
        }}
      >
        30 Days
      </ZButton>
    </div>

    <p class="text-right text-sm text-gray-500">
      {md(data.timeslots[0])} - {md(data.timeslots[data.timeslots.length - 1])}
    </p>
  </div>

  {#if loading}
    <div class="flex h-20 items-center justify-center">
      <ZSpinner />
    </div>
  {:else if failed}
    <div class="rounded border border-red-300 bg-red-50 p-4 text-sm text-red-700">조회 실패: {failed}</div>
  {:else}
    <section>
      <p class="mb-2 text-gray-500">Cloudflare Analytics</p>
      {#each rows as row, idx (row.key)}
        <div class="grid items-center md:grid-cols-[220px_minmax(0,760px)] md:justify-center">
          <aside class="rounded">
            <div class="text-gray-500">{row.label}</div>
            <div class="text-[1.2rem] font-bold">{row.statText}</div>
          </aside>

          <LineChart
            title={row.label}
            {labels}
            unit={row.unit}
            selectedLabelMode={range === '24h' ? 'hour' : 'date'}
            hoveredIndex={syncedHoverIndex}
            onHoverIndex={(index) => {
              syncedHoverIndex = index
            }}
            series={row.series}
          />
        </div>
        {#if idx < rows.length - 1}
          <hr class="border-0 border-t border-gray-200 dark:border-gray-700" />
        {/if}
      {/each}
    </section>

    {#if range !== '24h'}
      <section class="mt-8">
        <p class="mb-2 text-gray-500">MediaWiki Statistics</p>
        {#each mwRows as row, idx (row.key)}
          <div class="grid items-center md:grid-cols-[220px_minmax(0,760px)] md:justify-center">
            <aside class="rounded">
              <div class="text-gray-500">{row.label}</div>
              <div class="text-[1.2rem] font-bold">{row.statText}</div>
            </aside>

            <LineChart
              title={row.label}
              labels={labelsMw}
              unit={row.unit}
              selectedLabelMode="date"
              hoveredIndex={syncedHoverIndex}
              onHoverIndex={(index) => {
                syncedHoverIndex = index
              }}
              series={row.series}
            />
          </div>
          {#if idx < mwRows.length - 1}
            <hr class="border-0 border-t border-gray-200 dark:border-gray-700" />
          {/if}
        {/each}
      </section>
    {/if}
  {/if}
</div>
