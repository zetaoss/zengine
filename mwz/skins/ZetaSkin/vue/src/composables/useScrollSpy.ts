// useScrollSpy.ts
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'

export function useScrollSpy(
  anchorsRef: { value: string[] },
  offset: number = 64,
) {
  const activeIds = ref<string[]>([])

  const getSections = () => {
    return anchorsRef.value
      .map(id => {
        const el = document.getElementById(id)
        if (!el) return null

        const rect = el.getBoundingClientRect()
        const top = rect.top + window.scrollY
        const bottom = rect.bottom + window.scrollY

        return { id, top, bottom }
      })
      .filter((x): x is { id: string; top: number; bottom: number } => !!x)
      .sort((a, b) => a.top - b.top)
  }

  const computeActiveIds = () => {
    const list = getSections()
    if (!list.length) {
      activeIds.value = []
      return
    }

    const viewportTop = window.scrollY + offset
    const viewportBottom = window.scrollY + window.innerHeight

    const visible = list.filter(
      s => s.bottom > viewportTop && s.top < viewportBottom,
    )

    if (visible.length) {
      activeIds.value = visible.map(s => s.id)
    } else {
      let current: string | null = null
      for (const s of list) {
        if (s.top <= viewportTop) current = s.id
        else break
      }
      if (!current) current = list[0]?.id ?? null
      if (!current) {
        activeIds.value = []
        return
      }
      activeIds.value = [current]
    }
  }

  let raf = 0
  const scheduleRecalc = () => {
    if (raf) return
    raf = requestAnimationFrame(() => {
      raf = 0
      computeActiveIds()
    })
  }

  onMounted(() => {
    scheduleRecalc()
    window.addEventListener('scroll', scheduleRecalc, { passive: true })
    window.addEventListener('resize', scheduleRecalc)
  })

  onBeforeUnmount(() => {
    if (raf) cancelAnimationFrame(raf)
    window.removeEventListener('scroll', scheduleRecalc)
    window.removeEventListener('resize', scheduleRecalc)
  })

  watch(() => anchorsRef.value, scheduleRecalc)

  return { activeIds }
}
