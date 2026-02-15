<!-- ResizableBox.svelte -->
<script lang="ts">
  import { onDestroy, onMount } from 'svelte'

  export let className = ''

  const minPx = 200

  let container: HTMLElement | null = null
  let marginPx = 0
  let isDragging = false
  let startX = 0
  let startMargin = 0
  let maxMargin = 0

  const onMouseMove = (e: MouseEvent) => {
    if (!isDragging || !container) return

    const rect = container.getBoundingClientRect()
    maxMargin = Math.max(0, rect.width - minPx)

    const delta = startX - e.clientX
    let next = startMargin + delta

    if (next < 0) next = 0
    if (next > maxMargin) next = maxMargin

    marginPx = next
  }

  const onMouseUp = () => {
    if (!isDragging) return
    isDragging = false
  }

  const onHandleDown = (e: MouseEvent) => {
    if (!container) return

    isDragging = true
    startX = e.clientX

    const rect = container.getBoundingClientRect()
    maxMargin = Math.max(0, rect.width - minPx)

    startMargin = marginPx
  }

  onMount(() => {
    window.addEventListener('mousemove', onMouseMove)
    window.addEventListener('mouseup', onMouseUp)
  })

  onDestroy(() => {
    window.removeEventListener('mousemove', onMouseMove)
    window.removeEventListener('mouseup', onMouseUp)
  })
</script>

<div bind:this={container} class={`relative w-full max-w-full ${className}`}>
  <div class="h-full pr-4" style={`margin-right: ${marginPx}px; box-sizing: border-box;`}>
    <slot />
  </div>

  {#if isDragging}
    <div class="absolute inset-0 z-10 cursor-col-resize"></div>
  {/if}

  <button
    type="button"
    class="absolute inset-y-0 z-20 flex w-4 cursor-col-resize select-none items-center justify-center bg-white hover:bg-slate-200 text-xs text-slate-400 rounded-r"
    style={`right: ${marginPx}px;`}
    aria-label="Resize output"
    on:mousedown|stopPropagation|preventDefault={onHandleDown}
  >
    ◀
  </button>
</div>
