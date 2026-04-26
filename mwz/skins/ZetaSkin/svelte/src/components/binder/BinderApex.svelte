<svelte:options customElement={{ tag: 'binder-apex', shadow: 'none' }} />

<script lang="ts">
  import { mdiMenu } from '@mdi/js'
  import { onMount } from 'svelte'

  import type { Binder } from '$lib/types/binder'
  import getRLCONF from '$lib/utils/rlconf'
  import ZIcon from '$shared/ui/ZIcon.svelte'

  import BinderNode from './BinderNode.svelte'

  function isBinder(value: unknown): value is Binder {
    if (!value || typeof value !== 'object') return false
    const obj = value as { id?: unknown; text?: unknown; nodes?: unknown }
    return typeof obj.id === 'number' && typeof obj.text === 'string' && Array.isArray(obj.nodes)
  }

  function toBinders(value: unknown): Binder[] {
    return Array.isArray(value) ? value.filter(isBinder) : []
  }

  const { wgArticleId, binders } = getRLCONF()

  let bindersRef: Binder[] = toBinders(binders)
  let isCollapsed = false
  let isDrawer = false
  let refreshingId: number | null = null
  let root: HTMLElement | null = null

  const updateMedia = () => {
    const match = window.matchMedia('(max-width: 768px)')
    isDrawer = match.matches
    if (isDrawer && !isCollapsed) isCollapsed = true
    if (!isDrawer && isCollapsed) isCollapsed = false
  }

  const toggle = () => {
    isCollapsed = !isCollapsed
  }

  const close = () => {
    isCollapsed = true
  }

  const onMouseDown = (e: MouseEvent) => {
    if (!isDrawer || isCollapsed) return
    if (!root) return
    if (root.contains(e.target as Node)) return
    close()
  }

  async function refreshBinder() {
    if (refreshingId !== null) return
    refreshingId = wgArticleId
    try {
      const res = await fetch(`/w/rest.php/binder/${wgArticleId}?refresh=1`)
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      await res.json()
      window.location.reload()
    } catch (e) {
      console.error(e)
    } finally {
      refreshingId = null
    }
  }

  const handleRefresh = async (event: MouseEvent) => {
    event.preventDefault()
    event.stopPropagation()
    await refreshBinder()
  }

  $: styleVars = (() => {
    const marginY = 0
    const top = `calc(var(--navbar-visible-height) + ${marginY}px)`
    const height = isDrawer
      ? `calc(100vh - (var(--navbar-visible-height) + ${marginY * 2}px))`
      : `calc(100vh - (var(--navbar-visible-height) + var(--footer-visible-height) + ${marginY * 2}px))`
    return {
      width: isCollapsed ? '0' : '240px',
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

{#if bindersRef.length > 0}
  <div
    bind:this={root}
    class={`flex-none shrink-0 z-30 transition-[width] bg-(--background-color-neutral-subtle) border-r ${isDrawer ? 'fixed' : 'sticky'}`}
    style={`width:${styleVars.width}; margin-top:${styleVars.marginTop}; top:${styleVars.top}; height:${styleVars.height};`}
  >
    {#if isDrawer}
      <button
        class="binder-menu-toggle absolute p-2 z-20 flex rounded-r opacity-80 cursor-pointer bg-(--background-color-interactive--hover) right-0 translate-x-full"
        aria-expanded={!isCollapsed}
        on:click={toggle}
      >
        <ZIcon path={mdiMenu} />
      </button>
    {/if}

    <div class="z-scrollbar h-full w-full overflow-y-auto">
      {#each bindersRef as binder (binder.id)}
        <div>
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <header
            class="book sticky top-0 z-10 flex items-center justify-between rounded-lg px-3 py-2 bg-gray-200/80 dark:bg-gray-800/80 border-gray-400/60 dark:border-gray-600/60 font-bold"
            on:dblclick={handleRefresh}
          >
            <a href={`/wiki/Binder:${binder.text}`} class="binder-title-link inline-flex min-w-0 items-center gap-2">
              <span class="truncate">{binder.text}</span>
              {#if refreshingId === wgArticleId}
                <span class="refresh-dot" aria-hidden="true"></span>
              {/if}
            </a>
          </header>

          <ul class="m-0 p-0 pt-2 pb-10 list-none text-[.9rem]">
            {#each binder.nodes || [] as node, index (`${binder.id}:${index}`)}
              <BinderNode {node} depth={0} {wgArticleId} binderId={binder.id} pathKey={`${binder.id}:${index}`} />
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
    position: absolute;
    bottom: 0;
    left: 0;
    width: calc(100% - 6px);
    height: 64px;
    background-color: var(--bg-muted);
    mask-image: linear-gradient(transparent, #000 64px);
    pointer-events: none;
  }

  .binder-menu-toggle:hover {
    background-color: rgb(209 213 219);
  }

  .binder-title-link {
    color: var(--color-base);
    text-decoration: none;
  }

  .binder-title-link:hover {
    color: var(--z-link);
    text-decoration: underline;
  }

  :global(.dark) .binder-menu-toggle:hover {
    background-color: rgb(82 82 82);
  }

  .refresh-dot {
    width: 0.45rem;
    height: 0.45rem;
    flex: 0 0 auto;
    border-radius: 9999px;
    animation: binder-refresh-dot 0.8s ease-in-out infinite;
    background-color: currentColor;
    opacity: 0.5;
  }

  @keyframes binder-refresh-dot {
    0%,
    100% {
      opacity: 0.2;
      transform: scale(0.8);
    }
    50% {
      opacity: 0.8;
      transform: scale(1);
    }
  }
</style>
