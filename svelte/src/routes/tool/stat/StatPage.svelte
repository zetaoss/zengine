<svelte:options runes={true} />

<script lang="ts">
  import { SvelteURL } from 'svelte/reactivity'

  import { goto } from '$app/navigation'
  import { resolve } from '$app/paths'
  import { page } from '$app/state'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import ZTabs from '$shared/ui/ZTabs.svelte'
  import ZToggle from '$shared/ui/ZToggle.svelte'
  import httpy from '$shared/utils/httpy'

  import LineChart from './LineChart.svelte'

  type MetricKey = 'uniq_uniques' | 'sum_requests' | 'sum_bytes' | 'sum_cachedBytes'
  type GaMetricKey = 'active_users' | 'screen_page_views' | 'sessions'
  type ChartUnit = 'count' | 'bytes' | 'cores' | 'percent' | 'rank'

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

  type K8sMetricKey =
    | 'node_cpu_usage'
    | 'node_cpu_allocatable'
    | 'node_memory_usage'
    | 'node_memory_allocatable'
    | 'pod_cpu_usage'
    | 'pod_memory_usage'
    | 'pvc_storage_usage'
    | 'pvc_storage_capacity'
    | 'pod_count'
    | 'defender_fighting_ratio'
    | 'defender_max_level'

  interface K8sResp {
    timeslots: string[]
    node_cpu_usage: Array<unknown>
    node_cpu_allocatable: Array<unknown>
    node_memory_usage: Array<unknown>
    node_memory_allocatable: Array<unknown>
    pod_cpu_usage: Array<unknown>
    pod_memory_usage: Array<unknown>
    pvc_storage_usage: Array<unknown>
    pvc_storage_capacity: Array<unknown>
    pod_count: Array<unknown>
    defender_fighting_ratio: Array<unknown>
    defender_max_level: Array<unknown>
  }

  interface RowDef {
    key: string
    label: string
    value: number | null
    secondaryValue?: number | null
    unit: ChartUnit
    series: Array<{ label: string; color?: string; area?: boolean; layer?: 'main' | 'sub'; values: Array<number | null> }>
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
  const EMPTY_K8S: K8sResp = {
    timeslots: [],
    node_cpu_usage: [],
    node_cpu_allocatable: [],
    node_memory_usage: [],
    node_memory_allocatable: [],
    pod_cpu_usage: [],
    pod_memory_usage: [],
    pvc_storage_usage: [],
    pvc_storage_capacity: [],
    pod_count: [],
    defender_fighting_ratio: [],
    defender_max_level: [],
  }
  const DEFAULT_LINE_COLOR = '#0891b2'
  const SUB_LINE_COLOR = 'var(--color-a-gray-300)'

  let loading = $state(true)
  let failed = $state<string | null>(null)

  function parseRange(value: string | undefined) {
    const r = value ?? ''
    if (r === '28d' || r === '180d') return r
    return '48h'
  }

  let range = $derived.by<'48h' | '28d' | '180d'>(() => parseRange(page.params.range))

  let observedRouteRange: string | null = null

  let valueMode = $state<'compact' | 'exact'>('compact')
  let diffModeByKey = $state<Record<string, boolean>>({})
  let syncedHoverIndex = $state<number | null>(null)
  let gscHoverIndex = $state<number | null>(null)
  let data = $state<AnalyticsResp>(EMPTY)
  let gaData = $state<GaResp>(EMPTY_GA)
  let gscData = $state<GscResp>(EMPTY_GSC)
  let mwData = $state<MwStatisticsResp>(EMPTY_MW)
  let k8sData = $state<K8sResp>(EMPTY_K8S)
  let fetchVersion = 0

  const rangeTabs = $derived.by(() => [
    { value: '48h', label: '48 Hours', href: resolve('/tool/stat/48h') },
    { value: '28d', label: '28 Days', href: resolve('/tool/stat/28d') },
    { value: '180d', label: '180 Days', href: resolve('/tool/stat/180d') },
  ])

  const labels = $derived.by(() => (range === '48h' ? data.timeslots : data.timeslots.map((v) => normalizeDateKey(v))))
  const labelsGa = $derived.by(() => (range === '48h' ? gaData.timeslots : gaData.timeslots.map((v) => normalizeDateKey(v))))
  const rows = $derived.by<RowDef[]>(() => buildRows(data))
  const gaRows = $derived.by<RowDef[]>(() => buildGaRows(gaData))
  const labelsGsc = $derived.by(() => (range === '48h' ? gscData.timeslots : gscData.timeslots.map((v) => normalizeDateKey(v))))
  const gscRows = $derived.by<RowDef[]>(() => buildGscRows(gscData))
  const gscSharesTimeAxis = $derived.by(() => sameLabels(labelsGsc, labels))
  const labelsMw = $derived.by(() => (range === '48h' ? mwData.timeslots : mwData.timeslots.map((v) => normalizeDateKey(v))))
  const mwRows = $derived.by<RowDef[]>(() => buildMwRows(mwData))
  const labelsK8s = $derived.by(() => (range === '48h' ? k8sData.timeslots : k8sData.timeslots.map((v) => normalizeDateKey(v))))
  const k8sRows = $derived.by<RowDef[]>(() => buildK8sRows(k8sData))
  const visibleTimeslots = $derived.by(() => {
    if (data.timeslots.length > 0) return data.timeslots
    if (gaData.timeslots.length > 0) return gaData.timeslots
    if (gscData.timeslots.length > 0) return gscData.timeslots
    if (k8sData.timeslots.length > 0) return k8sData.timeslots
    return mwData.timeslots
  })

  async function fetchData(selectedRange: '48h' | '28d' | '180d') {
    const version = ++fetchVersion
    loading = true
    failed = null

    if (selectedRange === '48h') {
      const [[cfResp, cfErr], [gaResp, gaErr], [gscResp, gscErr], [mwResp, mwErr], [k8sResp, k8sErr]] = await Promise.all([
        httpy.get<AnalyticsResp>('/api/stat/cf-analytics/hourly'),
        httpy.get<GaResp>('/api/stat/ga/hourly'),
        httpy.get<GscResp>('/api/stat/gsc/hourly'),
        httpy.get<MwStatisticsResp>('/api/stat/mw-statistics/hourly'),
        httpy.get<K8sResp>('/api/stat/k8s/hourly'),
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
      if (k8sErr) {
        failed = k8sErr.message
        loading = false
        return
      }
      data = normalizeResp(cfResp)
      gaData = normalizeGaResp(gaResp)
      gscData = normalizeGscResp(gscResp)
      mwData = normalizeMwResp(mwResp)
      k8sData = normalizeK8sResp(k8sResp)
      loading = false
      return
    }

    const days = selectedRange === '28d' ? 28 : 180
    const [[cfResp, cfErr], [gaResp, gaErr], [gscResp, gscErr], [mwResp, mwErr], [k8sResp, k8sErr]] = await Promise.all([
      httpy.get<AnalyticsResp>(`/api/stat/cf-analytics/daily/${days}`),
      httpy.get<GaResp>(`/api/stat/ga/daily/${days}`),
      httpy.get<GscResp>(`/api/stat/gsc/daily/${days}`),
      httpy.get<MwStatisticsResp>(`/api/stat/mw-statistics/daily/${days}`),
      httpy.get<K8sResp>(`/api/stat/k8s/daily/${days}`),
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
    if (k8sErr) {
      failed = k8sErr.message
      loading = false
      return
    }

    data = normalizeResp(cfResp)
    gaData = normalizeGaResp(gaResp)
    gscData = normalizeGscResp(gscResp)
    mwData = normalizeMwResp(mwResp)
    k8sData = normalizeK8sResp(k8sResp)
    loading = false
  }

  function setRange(nextRange: string) {
    const p = `/tool/stat/${nextRange}`
    const url = new SvelteURL(page.url)
    url.pathname = p
    void goto(resolve((url.pathname + url.search) as '/tool/stat'), { replaceState: true, noScroll: true })
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

  function normalizeK8sResp(input: K8sResp | null): K8sResp {
    if (!input) return EMPTY_K8S
    return {
      timeslots: Array.isArray(input.timeslots) ? input.timeslots.map((v) => String(v)) : [],
      node_cpu_usage: Array.isArray(input.node_cpu_usage) ? input.node_cpu_usage : [],
      node_cpu_allocatable: Array.isArray(input.node_cpu_allocatable) ? input.node_cpu_allocatable : [],
      node_memory_usage: Array.isArray(input.node_memory_usage) ? input.node_memory_usage : [],
      node_memory_allocatable: Array.isArray(input.node_memory_allocatable) ? input.node_memory_allocatable : [],
      pod_cpu_usage: Array.isArray(input.pod_cpu_usage) ? input.pod_cpu_usage : [],
      pod_memory_usage: Array.isArray(input.pod_memory_usage) ? input.pod_memory_usage : [],
      pvc_storage_usage: Array.isArray(input.pvc_storage_usage) ? input.pvc_storage_usage : [],
      pvc_storage_capacity: Array.isArray(input.pvc_storage_capacity) ? input.pvc_storage_capacity : [],
      pod_count: Array.isArray(input.pod_count) ? input.pod_count : [],
      defender_fighting_ratio: Array.isArray(input.defender_fighting_ratio) ? input.defender_fighting_ratio : [],
      defender_max_level: Array.isArray(input.defender_max_level) ? input.defender_max_level : [],
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

  function buildK8sRows(resp: K8sResp): RowDef[] {
    return [
      {
        key: 'k8s-node-cpu',
        label: 'Nodepool CPU',
        value: lastK8sMetric(resp, 'node_cpu_usage'),
        secondaryValue: lastK8sMetric(resp, 'node_cpu_allocatable'),
        unit: 'cores',
        series: [
          { label: 'Nodepool CPU Usage', area: true, layer: 'main', values: seriesOfK8s(resp, 'node_cpu_usage') },
          { label: 'Nodepool CPU Allocatable', color: SUB_LINE_COLOR, area: true, layer: 'sub', values: seriesOfK8s(resp, 'node_cpu_allocatable') },
        ],
      },
      {
        key: 'k8s-node-memory',
        label: 'Nodepool Memory',
        value: lastK8sMetric(resp, 'node_memory_usage'),
        secondaryValue: lastK8sMetric(resp, 'node_memory_allocatable'),
        unit: 'bytes',
        series: [
          { label: 'Nodepool Memory Usage', area: true, layer: 'main', values: seriesOfK8s(resp, 'node_memory_usage') },
          { label: 'Nodepool Memory Allocatable', color: SUB_LINE_COLOR, area: true, layer: 'sub', values: seriesOfK8s(resp, 'node_memory_allocatable') },
        ],
      },
      {
        key: 'k8s-pod-cpu',
        label: 'Namespace CPU',
        value: lastK8sMetric(resp, 'pod_cpu_usage'),
        unit: 'cores',
        series: [{ label: 'Namespace CPU Usage', values: seriesOfK8s(resp, 'pod_cpu_usage') }],
      },
      {
        key: 'k8s-pod-memory',
        label: 'Namespace Memory',
        value: lastK8sMetric(resp, 'pod_memory_usage'),
        unit: 'bytes',
        series: [{ label: 'Namespace Memory Usage', values: seriesOfK8s(resp, 'pod_memory_usage') }],
      },
      {
        key: 'k8s-pvc-storage',
        label: 'PVC Storage',
        value: lastK8sMetric(resp, 'pvc_storage_usage'),
        secondaryValue: lastK8sMetric(resp, 'pvc_storage_capacity'),
        unit: 'bytes',
        series: [
          { label: 'PVC Storage Usage', area: true, layer: 'main', values: seriesOfK8s(resp, 'pvc_storage_usage') },
          { label: 'PVC Storage Capacity', color: SUB_LINE_COLOR, area: true, layer: 'sub', values: seriesOfK8s(resp, 'pvc_storage_capacity') },
        ],
      },
      {
        key: 'defender-fighting-ratio',
        label: 'Defender Fighting Time',
        value: defenderFightingPercent(resp),
        unit: 'percent',
        series: [{ label: 'Fighting Ratio', values: defenderFightingPercentSeries(resp) }],
      },
      {
        key: 'defender-max-level',
        label: 'Defender Max Level',
        value: lastK8sMetric(resp, 'defender_max_level'),
        unit: 'count',
        series: [{ label: 'Highest Defender Level', values: seriesOfK8s(resp, 'defender_max_level') }],
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

  function seriesOfK8s(resp: K8sResp, key: K8sMetricKey): Array<number | null> {
    const src = resp[key] ?? []
    return resp.timeslots.map((_, idx) => toNumber(src[idx] ?? null))
  }

  function lastK8sMetric(resp: K8sResp, key: K8sMetricKey): number | null {
    const src = resp[key] ?? []
    for (let i = src.length - 1; i >= 0; i -= 1) {
      const value = toNumber(src[i] ?? null)
      if (value != null) return value
    }
    return null
  }

  function defenderFightingPercent(resp: K8sResp): number | null {
    const ratio = lastK8sMetric(resp, 'defender_fighting_ratio')
    return ratio == null ? null : ratio * 100
  }

  function defenderFightingPercentSeries(resp: K8sResp): Array<number | null> {
    return seriesOfK8s(resp, 'defender_fighting_ratio').map((ratio) => (ratio == null ? null : ratio * 100))
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
    if (!Number.isInteger(value)) return `${stripZero(value.toFixed(1))}`
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
    if (!Number.isInteger(value)) {
      return value.toLocaleString('en-US', { maximumFractionDigits: 4 })
    }
    return value.toLocaleString('en-US', { maximumFractionDigits: 0 })
  }

  function formatStatValue(value: number | null, unit: ChartUnit) {
    if (unit === 'cores') return value == null ? '-' : `${Math.round(value * 1000)}m`
    if (valueMode === 'exact') return fmtExact(value, unit)
    if (unit === 'bytes') return fmtBytes(value)
    if (unit === 'percent') return fmtPercent(value)
    if (unit === 'rank') return fmtRank(value)
    return fmtCount(value)
  }

  function formatUsagePercent(usage: number | null, limit: number | null | undefined) {
    if (usage == null || limit == null || limit <= 0) return '-'
    const percentage = (usage / limit) * 100
    return `${percentage.toFixed(1)}%`
  }

  function stripZero(value: string) {
    if (value.includes('.')) {
      return value.replace(/\.?0+$/, '')
    }
    return value
  }

  function normalizeDateKey(value: string) {
    return value.trim().slice(0, 10)
  }

  function sameLabels(a: string[], b: string[]) {
    if (a.length === 0 || b.length === 0 || a.length !== b.length) return false
    for (let i = 0; i < a.length; i += 1) {
      if (a[i] !== b[i]) return false
    }
    return true
  }

  $effect(() => {
    if (range !== observedRouteRange) {
      const nextRange = range
      observedRouteRange = range
      syncedHoverIndex = null
      gscHoverIndex = null
      void fetchData(nextRange)
    }
  })
</script>

<div class="px-2 py-5">
  <h2 class="m-0 text-2xl font-bold">통계</h2>

  <div class="mb-4 mt-3 flex flex-wrap items-center justify-between gap-3">
    <ZTabs tabs={rangeTabs} selected={range} onChange={setRange} />

    <p class="text-sm text-a-gray-500">
      {md(visibleTimeslots[0])} - {md(visibleTimeslots[visibleTimeslots.length - 1])}
    </p>
  </div>

  {#if loading}
    <div class="flex h-20 items-center justify-center">
      <ZSpinner />
    </div>
  {:else if failed}
    <div class="rounded border border-a-red-300 bg-a-red-50 p-4 text-sm text-a-red-700">조회 실패: {failed}</div>
  {:else}
    <section>
      <p class="mb-2 text-a-gray-500">Cloudflare Analytics</p>
      {#each rows as row, idx (row.key)}
        <div class="grid items-center md:grid-cols-[180px_minmax(0,1fr)]">
          <aside class="rounded">
            <div class="text-a-gray-500">{row.label}</div>
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
          <hr class="border-0 border-t border-a-gray-200" />
        {/if}
      {/each}
    </section>

    <section class="mt-8">
      <p class="mb-2 text-a-gray-500">Google Analytics</p>
      {#each gaRows as row, idx (row.key)}
        <div class="grid items-center md:grid-cols-[180px_minmax(0,1fr)]">
          <aside class="rounded">
            <div class="text-a-gray-500">{row.label}</div>
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
          <hr class="border-0 border-t border-a-gray-200" />
        {/if}
      {/each}
    </section>

    <section class="mt-8">
      <p class="mb-2 text-a-gray-500">Google Search Console</p>
      {#each gscRows as row, idx (row.key)}
        <div class="grid items-center md:grid-cols-[180px_minmax(0,1fr)]">
          <aside class="rounded">
            <div class="text-a-gray-500">{row.label}</div>
            <div class="text-[1.2rem] font-bold">{formatStatValue(row.value, row.unit)}</div>
          </aside>

          <LineChart
            title={row.label}
            labels={labelsGsc}
            unit={row.unit}
            color={DEFAULT_LINE_COLOR}
            {valueMode}
            selectedLabelMode={range === '48h' ? 'hour' : 'date'}
            hoveredIndex={gscSharesTimeAxis ? syncedHoverIndex : gscHoverIndex}
            onHoverIndex={(index) => {
              if (gscSharesTimeAxis) syncedHoverIndex = index
              else gscHoverIndex = index
            }}
            series={row.series}
          />
        </div>
        {#if idx < gscRows.length - 1}
          <hr class="border-0 border-t border-a-gray-200" />
        {/if}
      {/each}
    </section>

    <section class="mt-8">
      <p class="mb-2 text-a-gray-500">MediaWiki Statistics</p>
      {#each mwRows as row, idx (row.key)}
        <div class="grid items-center md:grid-cols-[180px_minmax(0,1fr)]">
          <aside class="rounded">
            <div class="text-a-gray-500">{row.label}</div>
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
          <hr class="border-0 border-t border-a-gray-200" />
        {/if}
      {/each}
    </section>

    <section class="mt-8">
      <p class="mb-2 text-a-gray-500">Prom Metrics</p>
      {#each k8sRows as row, idx (row.key)}
        <div class="grid items-center md:grid-cols-[180px_minmax(0,1fr)]">
          <aside class="rounded">
            <div class="text-a-gray-500">{row.label}</div>
            <div class="text-[1.2rem] font-bold">
              {formatStatValue(row.value, row.unit)}
              {#if row.secondaryValue !== undefined}
                <span class="text-[0.9rem] font-medium text-a-gray-500">
                  / {formatStatValue(row.secondaryValue, row.unit)} ({formatUsagePercent(row.value, row.secondaryValue)})
                </span>
              {/if}
            </div>
          </aside>

          <LineChart
            title={row.label}
            labels={labelsK8s}
            unit={row.unit}
            color={DEFAULT_LINE_COLOR}
            {valueMode}
            fillArea={row.series.length === 1}
            showRatioPercent={row.secondaryValue !== undefined}
            selectedLabelMode={range === '48h' ? 'hour' : 'date'}
            hoveredIndex={syncedHoverIndex}
            onHoverIndex={(index) => {
              syncedHoverIndex = index
            }}
            series={row.series}
          />
        </div>
        {#if idx < k8sRows.length - 1}
          <hr class="border-0 border-t border-a-gray-200" />
        {/if}
      {/each}
    </section>

    <div class="mt-8 flex justify-center">
      <div class="flex flex-wrap items-center justify-center gap-6 text-sm text-a-gray-500">
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
