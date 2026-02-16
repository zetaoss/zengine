<svelte:options customElement={{ tag: 'toc-apex', shadow: 'none' }} />

<script lang="ts">
  import { onDestroy, onMount } from 'svelte'

  import type { Section } from '$lib/types/toc'
  import getRLCONF from '$lib/utils/rlconf'

  import TocTree from './TocTree.svelte'

  export let headerOffset: number | string = 64
  export let side: boolean | string = false

  let dataToc: unknown = getRLCONF().dataToc

  let tocRef: HTMLElement | null = null
  let isTocPast = false
  let activeIds: string[] = []
  let raf = 0

  const coerceBool = (value: boolean | string) => value === true || value === '' || value === 'true'

  const coerceNumber = (value: number | string, fallback: number) => {
    if (typeof value === 'number') return value
    const n = Number(value)
    return Number.isFinite(n) ? n : fallback
  }

  $: isSide = coerceBool(side)
  $: headerOffsetNum = coerceNumber(headerOffset, 64)

  const tocObj = typeof dataToc === 'object' && dataToc !== null ? (dataToc as Section) : null

  const flattenAnchors = (root?: Section | null): string[] => {
    if (!root) return []
    const out: string[] = []
    const walk = (n: Section) => {
      if (n.anchor) out.push(n.anchor)
      ;(n['array-sections'] ?? []).forEach(walk)
    }
    walk(root)
    return out
  }

  const anchors = flattenAnchors(tocObj)
  $: isDrawerMode = !isSide && isTocPast
  $: enableScrollSpy = isSide || isDrawerMode

  const getSections = () => {
    return anchors
      .map((id) => {
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
    if (!enableScrollSpy) {
      activeIds = []
      return
    }

    const list = getSections()
    if (!list.length) {
      activeIds = []
      return
    }

    const viewportTop = window.scrollY + headerOffsetNum
    const viewportBottom = window.scrollY + window.innerHeight

    const visible = list.filter((s) => s.bottom > viewportTop && s.top < viewportBottom)

    if (visible.length) {
      activeIds = visible.map((s) => s.id)
    } else {
      let current: string | null = null
      for (const s of list) {
        if (s.top <= viewportTop) current = s.id
        else break
      }
      if (!current) current = list[0]?.id ?? null
      if (!current) {
        activeIds = []
        return
      }
      activeIds = [current]
    }
  }

  const scheduleRecalc = () => {
    if (!enableScrollSpy) return
    if (raf) return
    raf = requestAnimationFrame(() => {
      raf = 0
      computeActiveIds()
    })
  }

  $: if (enableScrollSpy) {
    scheduleRecalc()
  } else {
    activeIds = []
  }

  $: if (enableScrollSpy) {
    scheduleRecalcForAnchors(anchors)
  }

  function scheduleRecalcForAnchors(anchorIds: string[]) {
    if (!anchorIds.length) {
      activeIds = []
      return
    }
    scheduleRecalc()
  }

  let observer: IntersectionObserver | null = null

  const observeToc = () => {
    if (observer) observer.disconnect()
    if (!tocRef) return
    observer = new IntersectionObserver(
      ([entry]) => {
        if (!entry) return
        isTocPast = entry.boundingClientRect.bottom <= headerOffsetNum
      },
      { rootMargin: `-${headerOffsetNum}px 0px 0px 0px`, threshold: 0 },
    )
    observer.observe(tocRef)
  }

  $: if (!isSide && 'IntersectionObserver' in window && tocRef) {
    observeToc()
  }

  $: if (isSide && observer) {
    observer.disconnect()
    observer = null
  }

  onMount(() => {
    dataToc = getRLCONF().dataToc
    scheduleRecalc()
    window.addEventListener('scroll', scheduleRecalc, { passive: true })
    window.addEventListener('resize', scheduleRecalc)

    return () => {
      if (raf) cancelAnimationFrame(raf)
      window.removeEventListener('scroll', scheduleRecalc)
      window.removeEventListener('resize', scheduleRecalc)
    }
  })

  onDestroy(() => {
    if (raf) cancelAnimationFrame(raf)
    if (observer) observer.disconnect()
  })

  const marginY = 16
  const styleVars = () => {
    const top = `calc(var(--navbar-visible-height) + ${marginY}px)`
    const marginTop = `${marginY}px`
    const height = `calc(100vh - (var(--navbar-visible-height) + var(--footer-visible-height) + ${marginY * 2}px))`
    return {
      marginTop,
      top,
      height,
    }
  }
</script>

{#if tocObj}
  {#if isSide}
    <div
      class="flex-none shrink-0 z-30 transition-[width] sticky"
      style={`position: sticky; margin-top:${styleVars().marginTop}; top:${styleVars().top}; height:${styleVars().height};`}
    >
      <div class="z-scrollbar h-full w-full overflow-y-auto">
        <TocTree toc={tocObj} {activeIds} headerOffset={headerOffsetNum} />
      </div>
    </div>
  {:else}
    <div bind:this={tocRef}>
      <div class="inline-block rounded-md border border-slate-200 p-3 my-3">
        <TocTree toc={tocObj} {activeIds} headerOffset={headerOffsetNum} showRail={false} />
      </div>
    </div>
  {/if}
{/if}

<style>
  :global(toc-apex) {
    display: contents;
  }
</style>
