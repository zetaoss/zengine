// useScrollSpy.ts
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'

export function useScrollSpy(
  anchorsRef: { value: string[] },
  offset: number = 64
) {
  const activeId = ref<string | null>(null)

  const getSections = () => {
    return anchorsRef.value
      .map((id) => {
        const el = document.getElementById(id)
        const top = el ? el.getBoundingClientRect().top + window.scrollY : Infinity
        return { id, top }
      })
      .filter((x) => x.top !== Infinity)
      .sort((a, b) => a.top - b.top)
  }

  const computeActiveId = () => {
    const list = getSections()
    if (!list.length) {
      activeId.value = null
      return
    }

    const y = window.scrollY + offset
    let current: string | null = null

    for (const s of list) {
      if (s.top <= y) current = s.id
      else break
    }
    if (!current) current = list[0].id

    activeId.value = current
  }

  let raf = 0
  const scheduleRecalc = () => {
    if (raf) return
    raf = requestAnimationFrame(() => {
      raf = 0
      computeActiveId()
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

  return { activeId }
}
