<svelte:options runes={true} />

<script lang="ts">
  import 'uplot/dist/uPlot.min.css'

  import { onDestroy, onMount } from 'svelte'
  import uPlot from 'uplot'

  type Unit = 'count' | 'bytes' | 'percent'

  interface LineSeries {
    label: string
    color: string
    values: Array<number | null>
  }

  interface Props {
    title: string
    labels: string[]
    unit: Unit
    series: LineSeries[]
    height?: number
    hoveredIndex?: number | null
    onHoverIndex?: (index: number | null) => void
    selectedLabelMode?: 'date' | 'hour'
  }

  const {
    title,
    labels,
    unit,
    series,
    height = 80,
    hoveredIndex = null,
    onHoverIndex,
    selectedLabelMode = 'date',
  }: Props = $props()

  let hostEl = $state<HTMLDivElement | null>(null)
  let chartEl = $state<HTMLDivElement | null>(null)
  let valueLabelEl = $state<HTMLDivElement | null>(null)
  let chart = $state<uPlot | null>(null)
  let resizeObserver = $state<ResizeObserver | null>(null)
  let width = $state(0)
  let renderedSignature = $state('')
  let applyingExternalSync = $state(false)
  let valueLabelX = $state(0)
  let valueLabelY = $state(0)

  const structureSignature = $derived.by(() => `${unit}::${series.map((s) => `${s.label}:${s.color}`).join('|')}`)

  const hasAnyData = $derived.by(() =>
    series.some((line) => line.values.some((v) => typeof v === 'number' && Number.isFinite(v))),
  )

  const selectedIdx = $derived.by(() => {
    if (hoveredIndex == null) return null
    if (hoveredIndex < 0 || hoveredIndex >= labels.length) return null
    return hoveredIndex
  })

  const valueLabelVisible = $derived.by(() => {
    if (selectedIdx == null) return false
    return series.some((line) => toNumber(line.values[selectedIdx] ?? null) !== null)
  })

  const selectedLabel = $derived.by(() => {
    if (selectedIdx == null) return '-'
    const label = labels[selectedIdx]
    return selectedLabelMode === 'hour' ? formatHourLabel(label) : formatDateLabel(label)
  })

  const tooltipRows = $derived.by(() => {
    if (selectedIdx == null) return [] as Array<{ line: LineSeries; value: number | null }>
    return series
      .map((line) => ({
        line,
        value: toNumber(line.values[selectedIdx] ?? null),
      }))
      .sort((a, b) => {
        const av = a.value == null ? Number.NEGATIVE_INFINITY : a.value
        const bv = b.value == null ? Number.NEGATIVE_INFINITY : b.value
        return bv - av
      })
  })

  function formatValue(value: number | null) {
    if (value == null) return '-'

    if (unit === 'percent') {
      return `${value.toFixed(1)}%`
    }

    if (unit === 'bytes') {
      const units = ['B', 'Ki', 'Mi', 'Gi', 'Ti']
      let n = value
      let u = 0
      while (n >= 1024 && u < units.length - 1) {
        n /= 1024
        u += 1
      }
      return `${n.toFixed(1)}${units[u]}`
    }

    const rounded = Math.round(value)
    const abs = Math.abs(rounded)
    if (abs >= 1_000_000_000) return `${(rounded / 1_000_000_000).toFixed(1)}B`
    if (abs >= 1_000_000) return `${(rounded / 1_000_000).toFixed(1)}M`
    if (abs >= 1_000) return `${(rounded / 1_000).toFixed(1)}k`
    return `${rounded}`
  }

  function formatDateLabel(value: string | undefined) {
    if (!value) return '-'
    const dayPart = value.trim().slice(0, 10)
    const compact = dayPart.replace(/[^0-9]/g, '')
    if (compact.length >= 8) {
      const m = Number(compact.slice(4, 6))
      const d = Number(compact.slice(6, 8))
      if (!Number.isFinite(m) || !Number.isFinite(d)) return '-'
      return `${m}.${d}`
    }
    return '-'
  }

  function formatHourLabel(value: string | undefined) {
    if (!value) return '-'
    const d = new Date(value)
    if (Number.isNaN(d.getTime())) return '-'
    return d.toLocaleTimeString(undefined, {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    })
  }

  function toNumber(value: number | null) {
    return typeof value === 'number' && Number.isFinite(value) ? value : null
  }

  function fillColor(color: string) {
    const alpha = 0.16

    if (/^#[0-9a-fA-F]{6}$/.test(color)) {
      const r = Number.parseInt(color.slice(1, 3), 16)
      const g = Number.parseInt(color.slice(3, 5), 16)
      const b = Number.parseInt(color.slice(5, 7), 16)
      return `rgba(${r}, ${g}, ${b}, ${alpha})`
    }
    if (/^#[0-9a-fA-F]{3}$/.test(color)) {
      const r = Number.parseInt(color[1] + color[1], 16)
      const g = Number.parseInt(color[2] + color[2], 16)
      const b = Number.parseInt(color[3] + color[3], 16)
      return `rgba(${r}, ${g}, ${b}, ${alpha})`
    }
    if (/^hsl\(/i.test(color)) return color.replace(/^hsl\(/i, 'hsla(').replace(/\)$/, `, ${alpha})`)
    if (/^rgb\(/i.test(color)) return color.replace(/^rgb\(/i, 'rgba(').replace(/\)$/, `, ${alpha})`)
    return color
  }

  function buildData(): uPlot.AlignedData {
    const x = labels.map((_, i) => i)
    const ys = series.map((line) => line.values.map((v) => (typeof v === 'number' && Number.isFinite(v) ? v : null)))
    return [x, ...ys]
  }

  function getWidth() {
    const base = chartEl?.clientWidth ?? hostEl?.clientWidth ?? 0
    return Math.max(Math.floor(base), 320)
  }

  function placeValueLabel(anchorLeft: number, anchorTop: number) {
    const chartH = chartEl?.clientHeight ?? height
    const chartOffsetLeft = chartEl?.offsetLeft ?? 0
    const chartOffsetTop = chartEl?.offsetTop ?? 0
    const containerW = hostEl?.clientWidth ?? width
    const containerH = hostEl?.clientHeight ?? chartH
    const labelW = valueLabelEl?.offsetWidth ?? 50
    const labelH = valueLabelEl?.offsetHeight ?? 28
    const gap = 14

    let x = chartOffsetLeft + anchorLeft - labelW / 2
    if (x + labelW > containerW - 4) x = containerW - labelW - 4
    if (x < 4) x = 4

    let y = chartOffsetTop + anchorTop - labelH - gap
    if (y < chartOffsetTop + 4) y = chartOffsetTop + anchorTop + gap
    if (y + labelH > containerH - 4) y = Math.max(chartOffsetTop + 4, containerH - labelH - 4)

    valueLabelX = x
    valueLabelY = y
  }

  function placeValueLabelForIndex(idx: number) {
    if (!chart) return
    if (idx < 0 || idx >= labels.length) return

    const anchorValue =
      series
        .map((line) => toNumber(line.values[idx] ?? null))
        .find((v): v is number => v !== null) ?? 0

    const x = chart.valToPos(idx, 'x')
    const y = chart.valToPos(anchorValue, 'y')
    placeValueLabel(x, y)
  }

  function updateHoverIndex(u: uPlot) {
    let idx = u.cursor.idx
    if (idx == null) {
      const cursorLeft = u.cursor.left
      if (typeof cursorLeft === 'number' && Number.isFinite(cursorLeft)) {
        const fromPos = Math.round(u.posToVal(cursorLeft, 'x'))
        idx = Number.isFinite(fromPos) ? fromPos : null
      }
    }

    if (idx == null || idx < 0 || idx >= labels.length) {
      if (!applyingExternalSync) onHoverIndex?.(null)
      return
    }

    if (!applyingExternalSync) {
      placeValueLabelForIndex(idx)
      onHoverIndex?.(idx)
    }
  }

  function createChart() {
    if (!chartEl) return
    const nextWidth = getWidth()
    if (nextWidth <= 0) return

    chart?.destroy()
    chart = null

    const opts: uPlot.Options = {
      width: nextWidth,
      height,
      legend: { show: false },
      scales: {
        x: { time: false },
        y: {
          range: (_, __, max) => {
            const upper = Number.isFinite(max) && max > 0 ? max : 1
            return [0, upper]
          },
        },
      },
      axes: [
        {
          show: true,
          grid: { show: true, stroke: 'rgba(120, 120, 120, 0.14)', width: 1 },
          values: (_, vals) => vals.map(() => ''),
          splits: () => {
            const len = labels.length
            if (len <= 0) return []
            if (len === 1) return [0]
            const ticks: number[] = []
            for (let i = 0; i < len; i += 1) ticks.push(i)
            return ticks
          },
          size: 0,
          ticks: { show: false },
        },
        {
          show: false,
          size: 0,
          ticks: { show: false },
        },
      ],
      series: [
        {},
        ...series.map((line) => ({
          stroke: line.color,
          fill: fillColor(line.color),
          width: 2.2,
          points: { show: true, size: 8, width: 0, fill: line.color, stroke: line.color },
        })),
      ],
      cursor: {
        drag: { x: false, y: false },
      },
      hooks: {
        setCursor: [
          (u) => {
            updateHoverIndex(u)
          },
        ],
      },
      padding: [8, 8, 8, 8],
    }

    chart = new uPlot(opts, buildData(), chartEl)
    renderedSignature = structureSignature
  }

  onMount(() => {
    width = getWidth()
    createChart()

    if (chartEl) {
      resizeObserver = new ResizeObserver(() => {
        const next = getWidth()
        if (next <= 0 || next === width) return
        width = next
        if (chart) chart.setSize({ width, height })
        else createChart()
      })
      resizeObserver.observe(chartEl)
    }
  })

  onDestroy(() => {
    resizeObserver?.disconnect()
    chart?.destroy()
  })

  $effect(() => {
    if (!chart) return
    if (renderedSignature !== structureSignature) {
      createChart()
      return
    }
    chart.setData(buildData())
  })

  $effect(() => {
    if (!chart) return
    if (hoveredIndex == null) {
      applyingExternalSync = true
      chart.setCursor({ left: -10, top: -10 }, false)
      applyingExternalSync = false
      return
    }
    if (hoveredIndex < 0 || hoveredIndex >= labels.length) return

    const left = chart.valToPos(hoveredIndex, 'x')
    const top = chart.cursor.top ?? Math.floor(height * 0.25)

    applyingExternalSync = true
    chart.setCursor({ left, top }, false)
    applyingExternalSync = false
    placeValueLabelForIndex(hoveredIndex)
  })
</script>

<div class="relative" bind:this={hostEl}>
  <div
    bind:this={chartEl}
    class="relative"
    aria-label={`${title} line chart`}
    role="img"
    onmouseleave={() => {
      onHoverIndex?.(null)
    }}
  ></div>

  {#if valueLabelVisible}
    <div
      bind:this={valueLabelEl}
      class="pointer-events-none absolute z-15 min-w-[78px] whitespace-nowrap text-center text-[15px] font-semibold leading-[1.2]"
      style={`left:${valueLabelX}px; top:${valueLabelY}px;`}
    >
      <span class="mr-1.5 font-medium text-gray-500">{selectedLabel}</span>
      {#if tooltipRows.length === 1}
        <span>{formatValue(tooltipRows[0]?.value ?? null)}</span>
      {:else}
        <span>{formatValue(tooltipRows[0]?.value ?? null)} / {formatValue(tooltipRows[1]?.value ?? null)}</span>
      {/if}
    </div>
  {/if}

  {#if !hasAnyData}
    <div class="py-2 text-center text-sm text-gray-500">No Data</div>
  {/if}
</div>

<style>
  div :global(.uplot) {
    width: 100%;
  }
</style>
