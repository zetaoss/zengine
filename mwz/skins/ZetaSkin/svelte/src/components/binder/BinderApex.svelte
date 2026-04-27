<svelte:options customElement={{ tag: 'binder-apex', shadow: 'none' }} />

<script lang="ts">
  import { mdiChevronLeft, mdiChevronRight, mdiMenu, mdiRefresh } from '@mdi/js'
  import { onMount } from 'svelte'

  import type { Binder } from '$lib/types/binder'
  import getRLCONF from '$lib/utils/rlconf'
  import ZButton from '$shared/ui/ZButton.svelte'
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

  const { wgArticleId, wgUserId, binders } = getRLCONF()

  let bindersRef: Binder[] = toBinders(binders)
  let isCollapsed = false
  let isDrawer = false
  let refreshingId: number | null = null
  let root: HTMLElement | null = null
  let scrollEl: HTMLElement | null = null
  let isScrolledToBottom = true

  const isLoggedIn = wgUserId > 0

  const updateMedia = () => {
    const isMobile = window.matchMedia('(max-width: 768px)').matches
    if (isDrawer !== isMobile) {
      isDrawer = isMobile
      isCollapsed = isMobile
      setTimeout(updateScrollState, 0)
    }
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

  const updateScrollState = () => {
    if (!scrollEl) {
      isScrolledToBottom = true
      return
    }

    isScrolledToBottom = scrollEl.scrollTop + scrollEl.clientHeight >= scrollEl.scrollHeight - 1
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

  $: styleVars = (() => {
    const marginY = 0
    const top = `calc(var(--navbar-visible-height) + ${marginY}px)`
    const height = isDrawer
      ? `calc(100vh - (var(--navbar-visible-height) + ${marginY * 2}px))`
      : `calc(100vh - (var(--navbar-visible-height) + var(--footer-visible-height) + ${marginY * 2}px))`
    return {
      width: isCollapsed ? (isDrawer ? '0' : '2.25rem') : '240px',
      marginTop: `${marginY}px`,
      top,
      height,
    }
  })()

  onMount(() => {
    updateMedia()
    updateScrollState()
    const media = window.matchMedia('(max-width: 768px)')
    const onChange = () => updateMedia()
    media.addEventListener('change', onChange)
    window.addEventListener('resize', updateMedia)
    window.addEventListener('resize', updateScrollState)
    document.addEventListener('mousedown', onMouseDown)

    return () => {
      media.removeEventListener('change', onChange)
      window.removeEventListener('resize', updateMedia)
      window.removeEventListener('resize', updateScrollState)
      document.removeEventListener('mousedown', onMouseDown)
    }
  })
</script>

{#if bindersRef.length > 0}
  <div
    bind:this={root}
    class={`flex-none shrink-0 z-30 transition-[width] bg-(--x-gray-50) border-r ${isDrawer ? 'fixed' : 'sticky'}`}
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

    {#if isCollapsed && !isDrawer}
      <div class="h-full w-full">
        <ZButton color="ghost" class="h-full! w-full! rounded-none! p-0!" title="바인더 펼치기" onclick={toggle}>
          <ZIcon path={mdiChevronRight} />
        </ZButton>
      </div>
    {:else}
      <div bind:this={scrollEl} class="z-scrollbar h-full w-full overflow-y-auto" on:scroll={updateScrollState}>
        {#each bindersRef as binder (binder.id)}
          <div>
            <header
              class="book sticky top-0 z-10 flex min-h-9 items-stretch overflow-hidden rounded bg-gray-200/80 dark:bg-gray-800/80 border-gray-400/60 dark:border-gray-600/60 font-bold"
            >
              <a href={`/wiki/Binder:${binder.text}`} class="binder-title-link inline-flex min-w-0 flex-1 items-center px-3 py-2">
                <span class="wrap-break-word">{binder.text}</span>
              </a>
              {#if isLoggedIn}
                <ZButton
                  color="ghost"
                  class="w-9! self-stretch rounded-none! p-0!"
                  disabled={refreshingId !== null}
                  title="바인더 새로고침"
                  onclick={refreshBinder}
                >
                  <ZIcon path={mdiRefresh} class={refreshingId === wgArticleId ? 'animate-spin' : ''} />
                </ZButton>
              {/if}
              {#if !isDrawer}
                <ZButton color="ghost" class="w-9! self-stretch rounded-none! p-0!" title="바인더 접기" onclick={toggle}>
                  <ZIcon path={mdiChevronLeft} />
                </ZButton>
              {/if}
            </header>

            <ul class="m-0 p-0 pt-2 pb-10 list-none text-[.9rem]">
              {#each binder.nodes || [] as node, index (`${binder.id}:${index}`)}
                <BinderNode {node} depth={0} {wgArticleId} binderId={binder.id} pathKey={`${binder.id}:${index}`} />
              {/each}
            </ul>
          </div>
        {/each}
      </div>
      {#if !isScrolledToBottom}
        <footer class="nav-blur"></footer>
      {/if}
    {/if}
  </div>
{/if}

<style>
  :global(binder-apex) {
    display: block;
  }

  .nav-blur {
    position: absolute;
    right: 6px;
    bottom: 0;
    left: 0;
    height: 50px;
    background: linear-gradient(to bottom, transparent, var(--x-gray-50) 55%);
    pointer-events: none;
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
</style>
