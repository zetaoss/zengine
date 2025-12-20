<!-- ContributionMap.vue -->
<script setup lang="ts">
import ZButton from '@common/ui/ZButton.vue'
import ZCard from '@common/ui/ZCard.vue'
import ZIcon from '@common/ui/ZIcon.vue'
import {
  mdiChevronDoubleLeft,
  mdiChevronDoubleRight,
  mdiChevronLeft,
  mdiChevronRight,
} from '@mdi/js'
import { useDateFormat } from '@vueuse/core'
import { computed, ref, watch } from 'vue'

const isAnimating = ref(false)

type StatsMap = Record<string, number>

const props = defineProps<{
  stats: StatsMap | null | undefined
  minDate: Date
  maxDate: Date
}>()

const WINDOW_DAYS = 52 * 7
const MS_PER_DAY = 24 * 3600 * 1000

const toDateOnly = (d: Date) => new Date(d.getFullYear(), d.getMonth(), d.getDate())

const addDays = (base: Date, delta: number) => {
  const d = new Date(base)
  d.setDate(d.getDate() + delta)
  return toDateOnly(d)
}

const diffDays = (a: Date, b: Date) => {
  const aa = toDateOnly(a).getTime()
  const bb = toDateOnly(b).getTime()
  return Math.round((bb - aa) / MS_PER_DAY)
}

const valueMap = computed(() => {
  const map = new Map<string, number>()
  const src = props.stats || {}
  for (const [k, v] of Object.entries(src)) {
    map.set(k, Number(v) || 0)
  }
  return map
})

const minDateOnly = computed(() => toDateOnly(props.minDate))
const maxDateOnly = computed(() => toDateOnly(props.maxDate))

const viewStart = ref<Date>(minDateOnly.value)
const viewEnd = ref<Date>(maxDateOnly.value)

const normalizeWindow = (startRaw: Date) => {
  const min = minDateOnly.value
  const max = maxDateOnly.value
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

const setupInitialWindow = () => {
  const max = maxDateOnly.value
  const approxStart = addDays(max, -WINDOW_DAYS + 1)
  const { start, end } = normalizeWindow(approxStart)
  viewStart.value = start
  viewEnd.value = end
}

watch(
  () => [props.minDate, props.maxDate],
  () => {
    setupInitialWindow()
  },
  { immediate: true },
)

const canGoPrev = computed(() => viewStart.value > minDateOnly.value)
const canGoNext = computed(() => viewEnd.value < maxDateOnly.value)

const withAnimation = (duration: number, fn: () => void) => {
  isAnimating.value = true
  fn()
  setTimeout(() => {
    isAnimating.value = false
  }, duration)
}

const goToMin = () => {
  withAnimation(200, () => {
    const { start, end } = normalizeWindow(minDateOnly.value)
    viewStart.value = start
    viewEnd.value = end
  })
}

const goToMax = () => {
  withAnimation(200, () => {
    const max = maxDateOnly.value
    const approxStart = addDays(max, -WINDOW_DAYS + 1)
    const { start, end } = normalizeWindow(approxStart)
    viewStart.value = start
    viewEnd.value = end
  })
}

const goPrev = () => {
  withAnimation(200, () => {
    const approxStart = addDays(viewStart.value, -WINDOW_DAYS)
    const { start, end } = normalizeWindow(approxStart)
    viewStart.value = start
    viewEnd.value = end
  })
}

const goNext = () => {
  withAnimation(200, () => {
    const approxStart = addDays(viewStart.value, WINDOW_DAYS)
    const { start, end } = normalizeWindow(approxStart)
    viewStart.value = start
    viewEnd.value = end
  })
}

interface DayCell {
  date: Date
  key: string
  value: number
  weekIndex: number
  weekday: number
}

const cells = computed<DayCell[]>(() => {
  const items: DayCell[] = []

  const start = toDateOnly(viewStart.value)
  const gridStart = addDays(start, -start.getDay())

  const end = toDateOnly(viewEnd.value)
  const gridEnd = addDays(end, 6 - end.getDay())

  let idx = 0
  const cur = new Date(gridStart)

  while (cur <= gridEnd) {
    const weekIndex = Math.floor(idx / 7)
    const weekday = cur.getDay()
    const key = useDateFormat(cur, 'YYYY-MM-DD').value
    const value = valueMap.value.get(key) ?? 0

    items.push({
      date: new Date(cur),
      key,
      value,
      weekIndex,
      weekday,
    })

    cur.setDate(cur.getDate() + 1)
    idx++
  }

  return items
})

const weekCount = computed(() =>
  cells.value.length ? cells.value[cells.value.length - 1].weekIndex + 1 : 0,
)

const thresholds = [1, 10, 20, 40]
const getLevel = (value: number) => {
  if (value <= 0) return 0
  if (value < thresholds[0]) return 1
  if (value < thresholds[1]) return 2
  if (value < thresholds[2]) return 3
  if (value < thresholds[3]) return 4
  return 4
}

const CELL_SIZE = 10
const CELL_GAP = 3

interface MonthLabel {
  name: string
  weekIndex: number
}

const monthLabels = computed<MonthLabel[]>(() => {
  const labels: MonthLabel[] = []
  const groups = new Map<string, DayCell[]>()

  for (const cell of cells.value) {
    const key = useDateFormat(cell.date, 'YYYY-MM').value
    if (!groups.has(key)) groups.set(key, [])
    groups.get(key)!.push(cell)
  }

  for (const [, group] of groups) {
    if (!group.length) continue

    const sundayInFirstWeek = group.find(
      c => c.weekday === 0 && c.date.getDate() < 22,
    )

    if (!sundayInFirstWeek) continue

    labels.push({
      name: useDateFormat(sundayInFirstWeek.date, "'YY.MM").value,
      weekIndex: sundayInFirstWeek.weekIndex,
    })
  }

  return labels
})
</script>

<template>
  <ZCard class="p-6 mt-4 mycard">
    <template #header>
      <header class="flex items-center justify-between">
        <div>편집 달력</div>
        <div class="flex h-8 space-x-px">
          <ZButton class="rounded-r-none" @click="goToMin" :disabled="!canGoPrev">
            <ZIcon :path="mdiChevronDoubleLeft" />
          </ZButton>
          <ZButton class="rounded-none" @click="goPrev" :disabled="!canGoPrev">
            <ZIcon :path="mdiChevronLeft" />
          </ZButton>
          <ZButton class="rounded-none" @click="goNext" :disabled="!canGoNext">
            <ZIcon :path="mdiChevronRight" />
          </ZButton>
          <ZButton class="rounded-l-none" @click="goToMax" :disabled="!canGoNext">
            <ZIcon :path="mdiChevronDoubleRight" />
          </ZButton>
        </div>
      </header>
    </template>

    <div class="overflow-x-auto">
      <div class="flex justify-center min-w-max" :class="{ 'opacity-50': isAnimating }">
        <div>
          <div class="relative h-4">
            <span v-for="m in monthLabels" :key="m.name" class="absolute text-[8px]"
              :style="{ left: `${m.weekIndex * (CELL_SIZE + CELL_GAP)}px` }">
              {{ m.name }}
            </span>
          </div>

          <div class="grid grid-flow-col" :style="{
            gridTemplateColumns: `repeat(${weekCount}, ${CELL_SIZE}px)`,
            gridTemplateRows: `repeat(7, ${CELL_SIZE}px)`,
            columnGap: `${CELL_GAP}px`,
            rowGap: `${CELL_GAP}px`,
          }">
            <div v-for="cell in cells" :key="cell.key" class="box-border cursor-pointer hover:opacity-80"
              :data-date="cell.key" :data-value="cell.value" :title="`${cell.key} · ${cell.value} edits`" :style="{
                width: `${CELL_SIZE}px`,
                height: `${CELL_SIZE}px`,
                backgroundColor: `var(--level${getLevel(cell.value)})`,
                borderWidth: '0.5px',
                borderRadius: `2px`,
              }" />
          </div>
        </div>
      </div>
    </div>
  </ZCard>
</template>

<style scoped>
.mycard {
  --level0: #eff2f5;
  --level1: #aceebb;
  --level2: #4ac26b;
  --level3: #2da44e;
  --level4: #116329;
}

.dark .mycard {
  --level0: #282828;
  --level1: #033a16;
  --level2: #196c2e;
  --level3: #2ea043;
  --level4: #56d364;
}
</style>
