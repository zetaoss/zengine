<svelte:options customElement={{ tag: 'binder-apex', shadow: 'none' }} />

<script lang="ts">
  import { mdiCog, mdiMenu } from '@mdi/js'
  import { onMount } from 'svelte'

  import type { Binder } from '$lib/types/binder'
  import getRLCONF from '$lib/utils/rlconf'
  import ZIcon from '$shared/ui/ZIcon.svelte'

  import BinderNode from './BinderNode.svelte'

  function isBinder(value: unknown): value is Binder {
    if (!value || typeof value !== 'object') return false
    const obj = value as { id?: unknown; title?: unknown; trees?: unknown }
    return typeof obj.id === 'number' && typeof obj.title === 'string' && Array.isArray(obj.trees)
  }

  function toBinders(value: unknown): Binder[] {
    return Array.isArray(value) ? value.filter(isBinder) : []
  }

  const { wgArticleId, binders } = getRLCONF()

  let bindersRef: Binder[] = toBinders(binders)
  let busy = false
  let collapsed = false
  let isOverlay = false
  let root: HTMLElement | null = null

  const updateMedia = () => {
    const match = window.matchMedia('(max-width: 768px)')
    isOverlay = match.matches
    if (isOverlay && !collapsed) collapsed = true
    if (!isOverlay && collapsed) collapsed = false
  }

  const toggle = () => {
    collapsed = !collapsed
  }

  const close = () => {
    collapsed = true
  }

  const onMouseDown = (e: MouseEvent) => {
    if (!isOverlay || collapsed) return
    if (!root) return
    if (root.contains(e.target as Node)) return
    close()
  }

  async function refreshBinder() {
    if (busy) return
    busy = true
    try {
      const res = await fetch(`/w/rest.php/binder/${wgArticleId}?refresh=1`)
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = await res.json()
      bindersRef = toBinders(data)
    } catch (e) {
      console.error(e)
    } finally {
      busy = false
    }
  }

  $: hasBinders = bindersRef.length > 0

  $: styleVars = (() => {
    const marginY = 0
    const top = `calc(var(--navbar-visible-height) + ${marginY}px)`
    const height = isOverlay
      ? `calc(100vh - (var(--navbar-visible-height) + ${marginY * 2}px))`
      : `calc(100vh - (var(--navbar-visible-height) + var(--footer-visible-height) + ${marginY * 2}px))`
    return {
      width: collapsed ? '0' : '240px',
      marginTop: `${marginY}px`,
      top,
      height,
    }
  })()

  onMount(() => {
    updateMedia()
    const media = window.matchMedia('(max-width: 768px)')
    const onChange = () => updateMedia()
    media.addEventListener('change', onChange)
    window.addEventListener('resize', updateMedia)
    document.addEventListener('mousedown', onMouseDown)

    return () => {
      media.removeEventListener('change', onChange)
      window.removeEventListener('resize', updateMedia)
      document.removeEventListener('mousedown', onMouseDown)
    }
  })
</script>

{#if hasBinders}
  <div
    bind:this={root}
    class={`flex-none shrink-0 z-30 transition-[width] ${isOverlay ? 'fixed' : 'sticky'} ${collapsed ? 'overflow-hidden' : ''}`}
    style={`width:${styleVars.width}; margin-top:${styleVars.marginTop}; top:${styleVars.top}; height:${styleVars.height};`}
  >
    {#if isOverlay}
      <button
        class="absolute p-2 z-20 flex rounded-r-lg opacity-75 bg-gray-200 hover:bg-gray-300 dark:bg-neutral-700 dark:hover:bg-neutral-600 right-0 translate-x-full"
        aria-expanded={!collapsed}
        on:click={toggle}
      >
        <ZIcon path={mdiMenu} />
      </button>
    {/if}

    <div class="z-scrollbar h-full w-full overflow-y-auto">
      {#each bindersRef as binder (binder.id)}
        <div>
          <header
            class="book sticky top-0 z-10 flex items-center justify-between px-3 py-2 bg-gray-200/80 dark:bg-gray-800/80 border-gray-400/60 dark:border-gray-600/60 font-bold"
            role="button"
            tabindex="0"
            aria-label="Refresh binder"
            on:dblclick|stopPropagation={refreshBinder}
            on:keydown={(e) => {
              if (e.key === 'Enter') refreshBinder()
            }}
          >
            <span>{binder.title}</span>
            <a
              href={`/wiki/Binder:${binder.title}`}
              class="inline-flex items-center gap-1 rounded-md px-2 py-1"
              on:click|stopPropagation={() => {}}
            >
              <ZIcon path={mdiCog} class={`h-4 w-4 ${busy ? 'animate-spin' : ''}`} />
            </a>
          </header>

          <ul class="m-0 p-0 pt-2 pb-10 list-none text-[.9rem]">
            {#each binder.trees || [] as tree (tree.text)}
              <BinderNode node={tree} depth={0} {wgArticleId} binderId={binder.id} />
            {/each}
          </ul>
        </div>
      {/each}
    </div>
    <footer class="nav-blur"></footer>
  </div>
{/if}

<style>
  :global(binder-apex) {
    display: block;
  }

  .nav-blur {
    width: calc(100% - 6px);
    mask-image: linear-gradient(transparent, #000 64px);
    bottom: 0;
    height: 64px;
    left: 0;
    pointer-events: none;
    position: absolute;
    background-color: var(--bg-muted);
  }
</style>
