<svelte:options runes={true} />

<script lang="ts">
  import { mdiChevronDoubleLeft, mdiChevronDoubleRight, mdiChevronLeft, mdiChevronRight } from '@mdi/js'

  import ZButton from '$shared/ui/ZButton.svelte'
  import ZCard from '$shared/ui/ZCard.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'

  type StatsMap = Record<string, number>

  let { stats = null, minDate, maxDate }: { stats: StatsMap | null | undefined; minDate: Date; maxDate: Date } = $props()

  const WINDOW_DAYS = 52 * 7
  const MS_PER_DAY = 24 * 3600 * 1000

  const thresholds = [1, 10, 20, 40]

  const CELL_SIZE = 10
  const CELL_GAP = 3

  let isAnimating = $state(false)
  let viewStart = $state(new Date())
  let viewEnd = $state(new Date())

  function toDateOnly(d: Date): Date {
    return new Date(d.getFullYear(), d.getMonth(), d.getDate())
  }

  function formatYmd(d: Date): string {
    const y = d.getFullYear()
    const m = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    return `${y}-${m}-${day}`
  }

  function formatYmLabel(d: Date): string {
    const y = String(d.getFullYear()).slice(-2)
    const m = String(d.getMonth() + 1).padStart(2, '0')
    return `'${y}.${m}`
  }

  function addDays(base: Date, delta: number): Date {
    return toDateOnly(new Date(base.getTime() + delta * MS_PER_DAY))
  }

  function diffDays(a: Date, b: Date): number {
    const aa = toDateOnly(a).getTime()
    const bb = toDateOnly(b).getTime()
    return Math.round((bb - aa) / MS_PER_DAY)
  }

  let valueMap = $derived.by(() => {
    const map: Record<string, number> = {}
    const src = stats || {}

    for (const [k, v] of Object.entries(src)) {
      map[k] = Number(v) || 0
    }

    return map
  })

  let minDateOnly = $derived(toDateOnly(minDate))
  let maxDateOnly = $derived(toDateOnly(maxDate))

  function normalizeWindow(startRaw: Date) {
    const min = minDateOnly
    const max = maxDateOnly
    const totalSpan = diffDays(min, max) + 1

    if (totalSpan <= WINDOW_DAYS) {
      return { start: min, end: max }
    }

    let s = toDateOnly(startRaw)

    const minStart = min
    const maxStart = addDays(max, -WINDOW_DAYS + 1)

    if (s < minStart) s = minStart
    if (s > maxStart) s = maxStart

    const e = addDays(s, WINDOW_DAYS - 1)
    return { start: s, end: e }
  }

  function setupInitialWindow() {
    const max = maxDateOnly
    const approxStart = addDays(max, -WINDOW_DAYS + 1)
    const { start, end } = normalizeWindow(approxStart)
    viewStart = start
    viewEnd = end
  }

  $effect(() => {
    minDate.getTime()
    maxDate.getTime()
    setupInitialWindow()
  })

  let canGoPrev = $derived(viewStart > minDateOnly)
  let canGoNext = $derived(viewEnd < maxDateOnly)

  function withAnimation(duration: number, fn: () => void) {
    isAnimating = true
    fn()
    setTimeout(() => {
      isAnimating = false
    }, duration)
  }

  function goToMin() {
    withAnimation(200, () => {
      const { start, end } = normalizeWindow(minDateOnly)
      viewStart = start
      viewEnd = end
    })
  }

  function goToMax() {
    withAnimation(200, () => {
      const max = maxDateOnly
      const approxStart = addDays(max, -WINDOW_DAYS + 1)
      const { start, end } = normalizeWindow(approxStart)
      viewStart = start
      viewEnd = end
    })
  }

  function goPrev() {
    withAnimation(200, () => {
      const approxStart = addDays(viewStart, -WINDOW_DAYS)
      const { start, end } = normalizeWindow(approxStart)
      viewStart = start
      viewEnd = end
    })
  }

  function goNext() {
    withAnimation(200, () => {
      const approxStart = addDays(viewStart, WINDOW_DAYS)
      const { start, end } = normalizeWindow(approxStart)
      viewStart = start
      viewEnd = end
    })
  }

  interface DayCell {
    date: Date
    key: string
    value: number
    weekIndex: number
    weekday: number
  }

  let cells = $derived.by(() => {
    const items: DayCell[] = []

    const start = toDateOnly(viewStart)
    const gridStart = addDays(start, -start.getDay())

    const end = toDateOnly(viewEnd)
    const gridEnd = addDays(end, 6 - end.getDay())

    let idx = 0
    for (let t = gridStart.getTime(); t <= gridEnd.getTime(); t += MS_PER_DAY) {
      const cur = new Date(t)
      const weekIndex = Math.floor(idx / 7)
      const weekday = cur.getDay()
      const key = formatYmd(cur)
      const value = valueMap[key] ?? 0

      items.push({
        date: cur,
        key,
        value,
        weekIndex,
        weekday,
      })
      idx++
    }

    return items
  })

  let weekCount = $derived(cells.length > 0 ? cells[cells.length - 1].weekIndex + 1 : 0)

  function getLevel(value: number): number {
    if (value <= 0) return 0
    if (value < (thresholds[0] ?? 1)) return 1
    if (value < (thresholds[1] ?? 10)) return 2
    if (value < (thresholds[2] ?? 20)) return 3
    if (value < (thresholds[3] ?? 40)) return 4
    return 4
  }

  interface MonthLabel {
    name: string
    weekIndex: number
  }

  let monthLabels = $derived.by(() => {
    const labels: MonthLabel[] = []
    const groups: Record<string, DayCell[]> = {}

    for (const cell of cells) {
      const key = `${cell.date.getFullYear()}-${String(cell.date.getMonth() + 1).padStart(2, '0')}`
      if (!groups[key]) groups[key] = []
      groups[key].push(cell)
    }

    for (const group of Object.values(groups)) {
      if (group.length === 0) continue

      const sundayInFirstWeek = group.find((c) => c.weekday === 0 && c.date.getDate() < 22)
      if (!sundayInFirstWeek) continue

      labels.push({
        name: formatYmLabel(sundayInFirstWeek.date),
        weekIndex: sundayInFirstWeek.weekIndex,
      })
    }

    return labels
  })
</script>

<div class="mycard">
  <ZCard class="mt-4 p-6">
    <svelte:fragment slot="header">
      <header class="flex items-center justify-between">
        <div>편집 달력</div>
        <div class="flex h-8 space-x-px">
          <ZButton class="rounded-r-none" cooldown={0} onclick={goToMin} disabled={!canGoPrev}>
            <ZIcon path={mdiChevronDoubleLeft} />
          </ZButton>
          <ZButton class="rounded-none" cooldown={0} onclick={goPrev} disabled={!canGoPrev}>
            <ZIcon path={mdiChevronLeft} />
          </ZButton>
          <ZButton class="rounded-none" cooldown={0} onclick={goNext} disabled={!canGoNext}>
            <ZIcon path={mdiChevronRight} />
          </ZButton>
          <ZButton class="rounded-l-none" cooldown={0} onclick={goToMax} disabled={!canGoNext}>
            <ZIcon path={mdiChevronDoubleRight} />
          </ZButton>
        </div>
      </header>
    </svelte:fragment>

    <div class="overflow-x-auto">
      <div class="flex min-w-max justify-center" class:opacity-50={isAnimating}>
        <div>
          <div class="relative h-4">
            {#each monthLabels as m (m.name)}
              <span class="absolute text-[8px]" style={`left: ${m.weekIndex * (CELL_SIZE + CELL_GAP)}px`}>
                {m.name}
              </span>
            {/each}
          </div>

          <div
            class="grid grid-flow-col"
            style={`grid-template-columns: repeat(${weekCount}, ${CELL_SIZE}px); grid-template-rows: repeat(7, ${CELL_SIZE}px); column-gap: ${CELL_GAP}px; row-gap: ${CELL_GAP}px;`}
          >
            {#each cells as cell (cell.key)}
              <div
                class="box-border cursor-pointer rounded-[2px] border-[0.5px] hover:opacity-80"
                data-date={cell.key}
                data-value={cell.value}
                title={`${cell.key} · ${cell.value} edits`}
                style={`width: ${CELL_SIZE}px; height: ${CELL_SIZE}px; background-color: var(--level${getLevel(cell.value)});`}
              ></div>
            {/each}
          </div>
        </div>
      </div>
    </div>
  </ZCard>
</div>

<style>
  .mycard {
    --level0: #eff2f5;
    --level1: #aceebb;
    --level2: #4ac26b;
    --level3: #2da44e;
    --level4: #116329;
  }

  :global(.dark) .mycard {
    --level0: #282828;
    --level1: #033a16;
    --level2: #196c2e;
    --level3: #2ea043;
    --level4: #56d364;
  }
</style>
