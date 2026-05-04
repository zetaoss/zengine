<svelte:options runes={true} />

<script lang="ts">
  import { onMount } from 'svelte'

  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import ZTabs from '$shared/ui/ZTabs.svelte'
  import ZToggle from '$shared/ui/ZToggle.svelte'
  import httpy from '$shared/utils/httpy'

  import LineChart from './LineChart.svelte'

  type MetricKey = 'uniq_uniques' | 'sum_requests' | 'sum_bytes' | 'sum_cachedBytes'
  type GaMetricKey = 'active_users' | 'screen_page_views' | 'sessions'
  type ChartUnit = 'count' | 'bytes' | 'percent' | 'rank'

  interface AnalyticsResp {
    timeslots: string[]
    uniq_uniques: Array<unknown>
    sum_requests: Array<unknown>
    sum_bytes: Array<unknown>
    sum_cachedBytes: Array<unknown>
  }

  interface GaResp {
    timeslots: string[]
    active_users: Array<unknown>
    screen_page_views: Array<unknown>
    sessions: Array<unknown>
  }

  type GscMetricKey = 'clicks' | 'impressions' | 'ctr' | 'position'
  interface GscResp {
    timeslots: string[]
    clicks: Array<unknown>
    impressions: Array<unknown>
    ctr: Array<unknown>
    position: Array<unknown>
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
    value: number | null
    unit: ChartUnit
    series: Array<{ label: string; color?: string; values: Array<number | null> }>
    diffValues?: Array<number | null>
  }

  const EMPTY: AnalyticsResp = {
    timeslots: [],
    uniq_uniques: [],
    sum_requests: [],
    sum_bytes: [],
    sum_cachedBytes: [],
  }
  const EMPTY_GA: GaResp = {
    timeslots: [],
    active_users: [],
    screen_page_views: [],
    sessions: [],
  }
  const EMPTY_GSC: GscResp = {
    timeslots: [],
    clicks: [],
    impressions: [],
    ctr: [],
    position: [],
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
  const DEFAULT_LINE_COLOR = '#0891b2'

  let loading = $state(true)
  let failed = $state<string | null>(null)
  let range = $state<'48h' | '15d' | '90d'>('48h')
  let valueMode = $state<'compact' | 'exact'>('compact')
  let diffModeByKey = $state<Record<string, boolean>>({})
  let syncedHoverIndex = $state<number | null>(null)
  let data = $state<AnalyticsResp>(EMPTY)
  let gaData = $state<GaResp>(EMPTY_GA)
  let gscData = $state<GscResp>(EMPTY_GSC)
  let mwData = $state<MwStatisticsResp>(EMPTY_MW)
  let fetchVersion = 0

  const rangeTabs: Array<{ value: '48h' | '15d' | '90d'; label: string }> = [
    { value: '48h', label: '48 Hours' },
    { value: '15d', label: '15 Days' },
    { value: '90d', label: '90 Days' },
  ]

  const labels = $derived.by(() => (range === '48h' ? data.timeslots : data.timeslots.map((v) => normalizeDateKey(v))))
  const labelsGa = $derived.by(() => (range === '48h' ? gaData.timeslots : gaData.timeslots.map((v) => normalizeDateKey(v))))
  const rows = $derived.by<RowDef[]>(() => buildRows(data))
  const gaRows = $derived.by<RowDef[]>(() => buildGaRows(gaData))
  const labelsGsc = $derived.by(() => (range === '48h' ? gscData.timeslots : gscData.timeslots.map((v) => normalizeDateKey(v))))
  const gscRows = $derived.by<RowDef[]>(() => buildGscRows(gscData))
  const labelsMw = $derived.by(() => (range === '48h' ? mwData.timeslots : mwData.timeslots.map((v) => normalizeDateKey(v))))
  const mwRows = $derived.by<RowDef[]>(() => buildMwRows(mwData))
  const visibleTimeslots = $derived.by(() => {
    if (data.timeslots.length > 0) return data.timeslots
    if (gaData.timeslots.length > 0) return gaData.timeslots
    if (gscData.timeslots.length > 0) return gscData.timeslots
    return mwData.timeslots
  })

  async function fetchData() {
    const version = ++fetchVersion
    loading = true
    failed = null

    if (range === '48h') {
      const [[cfResp, cfErr], [gaResp, gaErr], [gscResp, gscErr], [mwResp, mwErr]] = await Promise.all([
        httpy.get<AnalyticsResp>('/api/stat/cf-analytics/hourly'),
        httpy.get<GaResp>('/api/stat/ga/hourly'),
        httpy.get<GscResp>('/api/stat/gsc/hourly'),
        httpy.get<MwStatisticsResp>('/api/stat/mw-statistics/hourly'),
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
      if (gscErr) {
        failed = gscErr.message
        loading = false
        return
      }
      if (gaErr) {
        failed = gaErr.message
        loading = false
        return
      }
      data = normalizeResp(cfResp)
      gaData = normalizeGaResp(gaResp)
      gscData = normalizeGscResp(gscResp)
      mwData = normalizeMwResp(mwResp)
      loading = false
      return
    }

    const days = range === '15d' ? 15 : 90
    const [[cfResp, cfErr], [gaResp, gaErr], [gscResp, gscErr], [mwResp, mwErr]] = await Promise.all([
      httpy.get<AnalyticsResp>(`/api/stat/cf-analytics/daily/${days}`),
      httpy.get<GaResp>(`/api/stat/ga/daily/${days}`),
      httpy.get<GscResp>(`/api/stat/gsc/daily/${days}`),
      httpy.get<MwStatisticsResp>(`/api/stat/mw-statistics/daily/${days}`),
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
    if (gscErr) {
      failed = gscErr.message
      loading = false
      return
    }
    if (gaErr) {
      failed = gaErr.message
      loading = false
      return
    }

    data = normalizeResp(cfResp)
    gaData = normalizeGaResp(gaResp)
    gscData = normalizeGscResp(gscResp)
    mwData = normalizeMwResp(mwResp)
    loading = false
  }

  function setRange(nextRange: string) {
    const value = nextRange as typeof range
    if (range === value) return
    range = value
    void fetchData()
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

  function normalizeGaResp(input: GaResp | null): GaResp {
    if (!input) return EMPTY_GA
    return {
      timeslots: Array.isArray(input.timeslots) ? input.timeslots.map((v) => String(v)) : [],
      active_users: Array.isArray(input.active_users) ? input.active_users : [],
      screen_page_views: Array.isArray(input.screen_page_views) ? input.screen_page_views : [],
      sessions: Array.isArray(input.sessions) ? input.sessions : [],
    }
  }

  function normalizeGscResp(input: GscResp | null): GscResp {
    if (!input) return EMPTY_GSC
    return {
      timeslots: Array.isArray(input.timeslots) ? input.timeslots.map((v) => String(v)) : [],
      clicks: Array.isArray(input.clicks) ? input.clicks : [],
      impressions: Array.isArray(input.impressions) ? input.impressions : [],
      ctr: Array.isArray(input.ctr) ? input.ctr : [],
      position: Array.isArray(input.position) ? input.position : [],
    }
  }

  function buildRows(resp: AnalyticsResp): RowDef[] {
    return [
      {
        key: 'unique-visitors',
        label: 'Unique Visitors',
        value: sumMetric(resp, 'uniq_uniques'),
        unit: 'count',
        series: [{ label: 'Unique Visitors', values: seriesOf(resp, 'uniq_uniques') }],
      },
      {
        key: 'total-requests',
        label: 'Total Requests',
        value: sumMetric(resp, 'sum_requests'),
        unit: 'count',
        series: [{ label: 'Total Requests', values: seriesOf(resp, 'sum_requests') }],
      },
      {
        key: 'percent-cached',
        label: 'Percent Cached',
        value: percentCachedTotal(resp),
        unit: 'percent',
        series: [{ label: 'Percent Cached', values: percentCachedSeries(resp) }],
      },
      {
        key: 'data-served',
        label: 'Total Data Served',
        value: sumMetric(resp, 'sum_bytes'),
        unit: 'bytes',
        series: [{ label: 'Total Data Served', values: seriesOf(resp, 'sum_bytes') }],
      },
      {
        key: 'data-cached',
        label: 'Data Cached',
        value: sumMetric(resp, 'sum_cachedBytes'),
        unit: 'bytes',
        series: [{ label: 'Data Cached', values: seriesOf(resp, 'sum_cachedBytes') }],
      },
    ]
  }

  function buildMwRows(resp: MwStatisticsResp): RowDef[] {
    return [
      {
        key: 'mw-pages',
        label: 'Pages',
        value: lastMetric(resp, 'pages'),
        unit: 'count',
        series: [{ label: 'Pages', values: seriesOfMw(resp, 'pages') }],
        diffValues: diffSeriesOfMw(resp, 'pages'),
      },
      {
        key: 'mw-articles',
        label: 'Articles',
        value: lastMetric(resp, 'articles'),
        unit: 'count',
        series: [{ label: 'Articles', values: seriesOfMw(resp, 'articles') }],
        diffValues: diffSeriesOfMw(resp, 'articles'),
      },
      {
        key: 'mw-edits',
        label: 'Edits',
        value: lastMetric(resp, 'edits'),
        unit: 'count',
        series: [{ label: 'Edits', values: seriesOfMw(resp, 'edits') }],
        diffValues: diffSeriesOfMw(resp, 'edits'),
      },
      {
        key: 'mw-images',
        label: 'Images',
        value: lastMetric(resp, 'images'),
        unit: 'count',
        series: [{ label: 'Images', values: seriesOfMw(resp, 'images') }],
        diffValues: diffSeriesOfMw(resp, 'images'),
      },
      {
        key: 'mw-users',
        label: 'Users',
        value: lastMetric(resp, 'users'),
        unit: 'count',
        series: [{ label: 'Users', values: seriesOfMw(resp, 'users') }],
      },
      {
        key: 'mw-activeusers',
        label: 'Active Users',
        value: lastMetric(resp, 'activeusers'),
        unit: 'count',
        series: [{ label: 'Active Users', values: seriesOfMw(resp, 'activeusers') }],
      },
      {
        key: 'mw-admins',
        label: 'Admins',
        value: lastMetric(resp, 'admins'),
        unit: 'count',
        series: [{ label: 'Admins', values: seriesOfMw(resp, 'admins') }],
      },
      {
        key: 'mw-jobs',
        label: 'Jobs',
        value: lastMetric(resp, 'jobs'),
        unit: 'count',
        series: [{ label: 'Jobs', values: seriesOfMw(resp, 'jobs') }],
      },
    ]
  }

  function buildGaRows(resp: GaResp): RowDef[] {
    return [
      {
        key: 'ga-active-users',
        label: 'Active Users',
        value: sumGaMetric(resp, 'active_users'),
        unit: 'count',
        series: [{ label: 'Active Users', values: seriesOfGa(resp, 'active_users') }],
      },
      {
        key: 'ga-views',
        label: 'Views',
        value: sumGaMetric(resp, 'screen_page_views'),
        unit: 'count',
        series: [{ label: 'Views', values: seriesOfGa(resp, 'screen_page_views') }],
      },
      {
        key: 'ga-sessions',
        label: 'Sessions',
        value: sumGaMetric(resp, 'sessions'),
        unit: 'count',
        series: [{ label: 'Sessions', values: seriesOfGa(resp, 'sessions') }],
      },
    ]
  }

  function buildGscRows(resp: GscResp): RowDef[] {
    return [
      {
        key: 'gsc-clicks',
        label: 'Clicks',
        value: sumGscMetric(resp, 'clicks'),
        unit: 'count',
        series: [{ label: 'Clicks', values: seriesOfGsc(resp, 'clicks') }],
      },
      {
        key: 'gsc-impressions',
        label: 'Impressions',
        value: sumGscMetric(resp, 'impressions'),
        unit: 'count',
        series: [{ label: 'Impressions', values: seriesOfGsc(resp, 'impressions') }],
      },
      {
        key: 'gsc-ctr',
        label: 'CTR',
        value: totalCtr(resp),
        unit: 'percent',
        series: [{ label: 'CTR', values: seriesOfGsc(resp, 'ctr') }],
      },
      {
        key: 'gsc-position',
        label: 'Position',
        value: weightedAveragePosition(resp),
        unit: 'rank',
        series: [{ label: 'Position', values: seriesOfGsc(resp, 'position') }],
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

  function seriesOfGa(resp: GaResp, key: GaMetricKey): Array<number | null> {
    const src = resp[key] ?? []
    return resp.timeslots.map((_, idx) => toNumber(src[idx] ?? null))
  }

  function sumGaMetric(resp: GaResp, key: GaMetricKey): number {
    let total = 0
    const src = resp[key] ?? []
    for (let i = 0; i < src.length; i += 1) {
      const value = toNumber(src[i] ?? null)
      if (value != null) total += value
    }
    return total
  }

  function seriesOfGsc(resp: GscResp, key: GscMetricKey): Array<number | null> {
    const src = resp[key] ?? []
    return resp.timeslots.map((_, idx) => toNumber(src[idx] ?? null))
  }

  function sumGscMetric(resp: GscResp, key: Extract<GscMetricKey, 'clicks' | 'impressions'>): number {
    let total = 0
    const src = resp[key] ?? []
    for (let i = 0; i < src.length; i += 1) {
      const value = toNumber(src[i] ?? null)
      if (value != null) total += value
    }
    return total
  }

  function totalCtr(resp: GscResp): number | null {
    const clicks = sumGscMetric(resp, 'clicks')
    const impressions = sumGscMetric(resp, 'impressions')
    if (!Number.isFinite(impressions) || impressions <= 0) return null
    return (clicks / impressions) * 100
  }

  function weightedAveragePosition(resp: GscResp): number | null {
    let weighted = 0
    let totalImpressions = 0

    for (let i = 0; i < resp.timeslots.length; i += 1) {
      const position = toNumber(resp.position[i] ?? null)
      const impressions = toNumber(resp.impressions[i] ?? null)
      if (position == null || impressions == null || impressions <= 0) continue
      weighted += position * impressions
      totalImpressions += impressions
    }

    if (!Number.isFinite(totalImpressions) || totalImpressions <= 0) return null
    return weighted / totalImpressions
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

  function diffSeriesOfMw(resp: MwStatisticsResp, key: MwMetricKey): Array<number | null> {
    const base = seriesOfMw(resp, key)
    return base.map((value, index) => {
      if (index === 0 || value == null) return null
      const prev = base[index - 1]
      if (prev == null) return null
      const diff = value - prev
      return Number.isFinite(diff) ? diff : null
    })
  }

  function supportsDiff(row: RowDef) {
    return Array.isArray(row.diffValues)
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

  function fmtRank(value: number | null | undefined) {
    if (typeof value !== 'number' || !Number.isFinite(value)) return '-'
    return value.toFixed(1)
  }

  function fmtExact(value: number | null | undefined, unit: ChartUnit) {
    if (typeof value !== 'number' || !Number.isFinite(value)) return '-'
    if (unit === 'percent') {
      return `${value.toLocaleString('en-US', { maximumFractionDigits: 4 })}%`
    }
    if (unit === 'bytes') {
      return `${value.toLocaleString('en-US', { maximumFractionDigits: 0 })} B`
    }
    if (unit === 'rank') {
      return value.toLocaleString('en-US', { minimumFractionDigits: 1, maximumFractionDigits: 2 })
    }
    return value.toLocaleString('en-US', { maximumFractionDigits: 0 })
  }

  function formatStatValue(value: number | null, unit: ChartUnit) {
    if (valueMode === 'exact') return fmtExact(value, unit)
    if (unit === 'bytes') return fmtBytes(value)
    if (unit === 'percent') return fmtPercent(value)
    if (unit === 'rank') return fmtRank(value)
    return fmtCount(value)
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

<div class="px-2 py-5">
  <h2 class="m-0 text-2xl font-bold">통계</h2>

  <div class="mb-4 mt-3 flex flex-wrap items-center justify-between gap-3">
    <ZTabs tabs={rangeTabs} selected={range} onChange={setRange} />

    <p class="text-sm text-gray-500 dark:text-gray-400">
      {md(visibleTimeslots[0])} - {md(visibleTimeslots[visibleTimeslots.length - 1])}
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
        <div class="grid items-center md:grid-cols-[180px_minmax(0,1fr)]">
          <aside class="rounded">
            <div class="text-gray-500">{row.label}</div>
            <div class="text-[1.2rem] font-bold">{formatStatValue(row.value, row.unit)}</div>
          </aside>

          <LineChart
            title={row.label}
            {labels}
            unit={row.unit}
            color={DEFAULT_LINE_COLOR}
            {valueMode}
            selectedLabelMode={range === '48h' ? 'hour' : 'date'}
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

    <section class="mt-8">
      <p class="mb-2 text-gray-500">Google Analytics</p>
      {#each gaRows as row, idx (row.key)}
        <div class="grid items-center md:grid-cols-[180px_minmax(0,1fr)]">
          <aside class="rounded">
            <div class="text-gray-500">{row.label}</div>
            <div class="text-[1.2rem] font-bold">{formatStatValue(row.value, row.unit)}</div>
          </aside>

          <LineChart
            title={row.label}
            labels={labelsGa}
            unit={row.unit}
            color={DEFAULT_LINE_COLOR}
            {valueMode}
            selectedLabelMode={range === '48h' ? 'hour' : 'date'}
            hoveredIndex={syncedHoverIndex}
            onHoverIndex={(index) => {
              syncedHoverIndex = index
            }}
            series={row.series}
          />
        </div>
        {#if idx < gaRows.length - 1}
          <hr class="border-0 border-t border-gray-200 dark:border-gray-700" />
        {/if}
      {/each}
    </section>

    <section class="mt-8">
      <p class="mb-2 text-gray-500">Google Search Console</p>
      {#each gscRows as row, idx (row.key)}
        <div class="grid items-center md:grid-cols-[180px_minmax(0,1fr)]">
          <aside class="rounded">
            <div class="text-gray-500">{row.label}</div>
            <div class="text-[1.2rem] font-bold">{formatStatValue(row.value, row.unit)}</div>
          </aside>

          <LineChart
            title={row.label}
            labels={labelsGsc}
            unit={row.unit}
            color={DEFAULT_LINE_COLOR}
            {valueMode}
            selectedLabelMode={range === '48h' ? 'hour' : 'date'}
            hoveredIndex={syncedHoverIndex}
            onHoverIndex={(index) => {
              syncedHoverIndex = index
            }}
            series={row.series}
          />
        </div>
        {#if idx < gscRows.length - 1}
          <hr class="border-0 border-t border-gray-200 dark:border-gray-700" />
        {/if}
      {/each}
    </section>

    <section class="mt-8">
      <p class="mb-2 text-gray-500">MediaWiki Statistics</p>
      {#each mwRows as row, idx (row.key)}
        <div class="grid items-center md:grid-cols-[180px_minmax(0,1fr)]">
          <aside class="rounded">
            <div class="text-gray-500">{row.label}</div>
            <div class="text-[1.2rem] font-bold">{formatStatValue(row.value, row.unit)}</div>
          </aside>

          <LineChart
            title={supportsDiff(row) && diffModeByKey[row.key] === true ? `${row.label} diff` : row.label}
            labels={labelsMw}
            unit={row.unit}
            color={DEFAULT_LINE_COLOR}
            {valueMode}
            fillArea={!(supportsDiff(row) && diffModeByKey[row.key] === true)}
            selectedLabelMode={range === '48h' ? 'hour' : 'date'}
            hoveredIndex={syncedHoverIndex}
            onHoverIndex={(index) => {
              syncedHoverIndex = index
            }}
            series={supportsDiff(row) && diffModeByKey[row.key] === true
              ? [{ label: `${row.label} diff`, values: row.diffValues ?? [] }]
              : row.series}
            barValues={supportsDiff(row) && diffModeByKey[row.key] === true ? (row.diffValues ?? []) : null}
            barColor={DEFAULT_LINE_COLOR}
          />
        </div>
        {#if idx < mwRows.length - 1}
          <hr class="border-0 border-t border-gray-200 dark:border-gray-700" />
        {/if}
      {/each}
    </section>

    <div class="mt-8 flex justify-center">
      <div class="flex flex-wrap items-center justify-center gap-6 text-sm text-gray-500 dark:text-gray-400">
        <div class="flex items-center gap-2">
          <span>exact</span>
          <ZToggle
            label="exact"
            checked={valueMode === 'exact'}
            showIcon={false}
            onchange={(event) => {
              valueMode = event.checked ? 'exact' : 'compact'
            }}
          />
        </div>

        <div class="flex items-center gap-2">
          <span>diff</span>
          <ZToggle
            label="diff"
            checked={Object.values(diffModeByKey).some(Boolean)}
            showIcon={false}
            onchange={(event) => {
              const checked = event.checked
              const next: Record<string, boolean> = {}
              for (const row of mwRows) {
                if (supportsDiff(row)) next[row.key] = checked
              }
              diffModeByKey = next
            }}
          />
        </div>
      </div>
    </div>
  {/if}
</div>
