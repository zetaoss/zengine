<script lang="ts">
  import type { Snippet } from 'svelte'
  import { tick } from 'svelte'

  interface Props {
    content?: string
    placement?: 'top' | 'bottom'
    ariaLabel?: string
    class?: string
    children?: Snippet
  }

  let { content = '', placement = 'top', ariaLabel = undefined, class: className = '', children }: Props = $props()

  let isOpen = $state(false)
  let isHovered = $state(false)
  let isPinned = $state(false)
  let triggerEl = $state<HTMLButtonElement | null>(null)
  let tooltipEl = $state<HTMLSpanElement | null>(null)
  let tooltipStyle = $state('')

  function updateTooltipPosition() {
    if (!triggerEl || !tooltipEl || typeof window === 'undefined') return

    const triggerRect = triggerEl.getBoundingClientRect()
    const tooltipRect = tooltipEl.getBoundingClientRect()
    const gap = 8
    const viewportMargin = 8

    const halfWidth = tooltipRect.width / 2
    const centerX = triggerRect.left + triggerRect.width / 2
    const left = Math.min(
      Math.max(centerX, viewportMargin + halfWidth),
      Math.max(viewportMargin + halfWidth, window.innerWidth - viewportMargin - halfWidth),
    )

    const preferredTop = placement === 'bottom' ? triggerRect.bottom + gap : triggerRect.top - tooltipRect.height - gap
    const flippedTop = placement === 'bottom' ? triggerRect.top - tooltipRect.height - gap : triggerRect.bottom + gap

    let top = preferredTop
    if (top < viewportMargin || top + tooltipRect.height > window.innerHeight - viewportMargin) {
      top = flippedTop
    }
    if (top < viewportMargin) top = viewportMargin
    if (top + tooltipRect.height > window.innerHeight - viewportMargin) {
      top = Math.max(viewportMargin, window.innerHeight - viewportMargin - tooltipRect.height)
    }

    tooltipStyle = `left: ${left}px; top: ${top}px;`
  }

  function openTooltip() {
    isHovered = true
    isOpen = true
    void tick().then(updateTooltipPosition)
  }

  function closeTooltip() {
    isHovered = false
    if (isPinned) return
    isOpen = false
  }

  function togglePinnedTooltip() {
    isPinned = !isPinned
    isOpen = isPinned || isHovered
    if (isOpen) {
      void tick().then(updateTooltipPosition)
    }
  }

  function closePinnedTooltip() {
    isPinned = false
    isOpen = isHovered
  }

  $effect(() => {
    void content
    void placement
    if (!isOpen) return

    void tick().then(updateTooltipPosition)

    const handleResize = () => updateTooltipPosition()
    const handleScroll = () => updateTooltipPosition()
    const handleKeyup = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && isPinned) {
        closePinnedTooltip()
      }
    }

    window.addEventListener('resize', handleResize)
    window.addEventListener('scroll', handleScroll, true)
    window.addEventListener('keyup', handleKeyup)

    return () => {
      window.removeEventListener('resize', handleResize)
      window.removeEventListener('scroll', handleScroll, true)
      window.removeEventListener('keyup', handleKeyup)
    }
  })
</script>

<span class={`group relative inline-flex items-center ${className}`.trim()}>
  <button
    bind:this={triggerEl}
    type="button"
    class="inline-flex cursor-pointer items-center border-0 bg-transparent p-0"
    aria-label={ariaLabel ?? content}
    onmouseenter={openTooltip}
    onmouseleave={closeTooltip}
    onfocus={openTooltip}
    onblur={closeTooltip}
    onclick={togglePinnedTooltip}
  >
    {@render children?.()}
  </button>
  {#if isPinned}
    <button
      type="button"
      class="fixed inset-0 z-40 cursor-default bg-x-slate-950/15 backdrop-saturate-0 backdrop-brightness-75"
      aria-label="툴팁 닫기"
      onclick={closePinnedTooltip}
    ></button>
  {/if}
  <span
    bind:this={tooltipEl}
    class={`pointer-events-none fixed z-50 w-[360px] max-w-[calc(100vw-1rem)] -translate-x-1/2 rounded border border-x-slate-300 bg-x-white px-2 py-1 text-xs text-x-slate-700 shadow-lg whitespace-normal wrap-break-word transition-opacity duration-150 ${isOpen ? 'visible opacity-100' : 'invisible opacity-0'}`}
    role="tooltip"
    style={tooltipStyle}
  >
    {content}
  </span>
</span>
