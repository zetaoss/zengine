<svelte:options customElement={{ tag: 'c-search', shadow: 'none' }} />

<script lang="ts">
  import { mdiFileDocumentOutline, mdiHistory, mdiMagnify, mdiShuffle, mdiSubdirectoryArrowRight, mdiTextSearch } from '@mdi/js'
  import { onDestroy, onMount } from 'svelte'

  import { useDismissable } from '$shared/composables/useDismissable'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import getShortcut from '$shared/utils/shortcut'

  interface Page {
    description: string | null
    excerpt: string
    id: number
    key: string
    matched_title: string | null
    thumbnail: Record<string, unknown> | null
    title: string
  }
  interface SearchResponse {
    pages?: Page[]
    [k: string]: unknown
  }

  let root: HTMLElement | null = null
  let input: HTMLInputElement | null = null
  let focused = false
  let expanded = false
  let keyword = ''
  let pages: Page[] = []

  let kIndex = -1
  let hIndex = -1

  let aborter: AbortController | null = null
  let debounceId: ReturnType<typeof setTimeout> | null = null

  const searchTitle = `검색 (${getShortcut('f')})`
  const randomTitle = `랜덤 (${getShortcut('x')})`
  const recentTitle = `바뀐글 (${getShortcut('r')})`

  $: displayQuery = kIndex >= 0 && kIndex < pages.length ? (pages[kIndex]?.title ?? keyword) : keyword

  $: currentIndex = hIndex >= 0 ? hIndex : kIndex

  function close() {
    expanded = false
    kIndex = -1
    hIndex = -1
  }

  useDismissable(() => root, {
    enabled: () => expanded,
    onDismiss: close,
  })

  async function fetchData(q: string) {
    const trimmed = q.trim()
    if (!trimmed) return

    aborter?.abort()
    const controller = new AbortController()
    aborter = controller

    try {
      const url = `/w/rest.php/v1/search/title?${new URLSearchParams({ q: trimmed, limit: '10' })}`
      const res = await fetch(url, { signal: controller.signal })
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const json = (await res.json()) as SearchResponse
      pages = Array.isArray(json.pages) ? json.pages : []
      expanded = true
      kIndex = -1
      hIndex = -1
    } catch (e: unknown) {
      if (e instanceof DOMException && e.name === 'AbortError') return
      console.error('[search]', e)
    } finally {
      if (aborter === controller) aborter = null
    }
  }

  function debouncedFetch(q: string) {
    if (debounceId) clearTimeout(debounceId)
    debounceId = setTimeout(() => fetchData(q), 400)
  }

  function onInput(e: Event) {
    const val = (e.target as HTMLInputElement).value
    keyword = val
    kIndex = -1
    hIndex = -1

    if (!val.trim()) {
      aborter?.abort()
      pages = []
      expanded = false
      return
    }
    debouncedFetch(val)
  }

  function onFocus() {
    focused = true
    if (keyword.trim()) expanded = true
  }

  function isTyping(el: EventTarget | null) {
    return el instanceof HTMLElement && (el.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(el.tagName))
  }

  // "/" focuses the search box unless the user is typing somewhere.
  function onWindowKeyDown(e: KeyboardEvent) {
    if (e.key !== '/' || e.ctrlKey || e.metaKey || e.altKey || e.isComposing || isTyping(e.target)) return
    e.preventDefault()
    input?.focus()
    input?.select()
  }

  onMount(() => {
    window.addEventListener('keydown', onWindowKeyDown)
    return () => window.removeEventListener('keydown', onWindowKeyDown)
  })

  function handleUpDown(offset: number) {
    const max = pages.length
    let next = kIndex + offset
    if (next < -1) next = max
    if (next > max) next = -1
    kIndex = next
  }

  function pageUrl(p: Page) {
    return `/wiki/${p.key.split('/').map(encodeURIComponent).join('/')}`
  }

  function goToSearch() {
    if (typeof window === 'undefined') return

    if (kIndex >= 0 && kIndex < pages.length) {
      const p = pages[kIndex]
      keyword = p.title
      close()
      window.location.href = pageUrl(p)
      return
    }

    const base = new URL('/w/index.php', window.location.origin)
    base.searchParams.set('title', '특수:검색')

    base.searchParams.set('search', displayQuery)
    if (kIndex === pages.length) {
      base.searchParams.set('fulltext', '1')
      base.searchParams.set('ns0', '1')
    }

    close()
    window.location.href = base.toString()
  }

  function onClick() {
    kIndex = -1
    goToSearch()
  }

  function onKeyDown(e: KeyboardEvent) {
    if (e.key === 'ArrowUp') {
      e.preventDefault()
      handleUpDown(-1)
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      handleUpDown(1)
    } else if (e.key === 'Enter') {
      goToSearch()
    } else if (e.key === 'Escape') {
      e.preventDefault()
      if (expanded) close()
      else input?.blur()
    }
  }

  function escapeHTML(s: string) {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;')
  }

  function highlight(needle: string, haystack: string) {
    const safeNeedle = needle.trim().replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    if (!safeNeedle) return escapeHTML(haystack)
    const re = new RegExp(safeNeedle, 'giu')
    return escapeHTML(haystack).replace(re, (s) => `<b>${escapeHTML(s)}</b>`)
  }

  onDestroy(() => {
    aborter?.abort()
  })
</script>

<div class="text-foreground ml-auto flex w-full md:max-w-2xl">
  <div class="m-1.5 grow">
    <div bind:this={root} class="relative" role="search">
      <div class={`relative flex h-9 border border-transparent bg-background rounded-t ${!expanded || !keyword.trim().length ? 'rounded-b' : ''}`}>
        <!-- Same column as the suggestion icons, so typed text lines up with suggestion text. -->
        <button type="button" class="z-10 flex h-full flex-none items-center bg-transparent pr-2 pl-3 text-muted-foreground hover:text-foreground focus:text-a-blue-700" aria-label="검색" on:click={onClick}>
          <ZIcon path={mdiMagnify} size={18} />
        </button>
        <!-- svelte-ignore a11y_accesskey -->
        <input
          accesskey="f"
          aria-label="search"
          type="search"
          class="grow h-full bg-transparent pr-3 outline-0"
          name="search"
          placeholder="검색..."
          title={searchTitle}
          autocomplete="off"
          value={displayQuery}
          bind:this={input}
          on:input={onInput}
          on:focus={onFocus}
          on:blur={() => (focused = false)}
          on:keydown={onKeyDown}
        />
        {#if !focused && !keyword}
          <kbd class="pointer-events-none absolute top-1/2 right-3 hidden -translate-y-1/2 text-muted-foreground md:inline" aria-hidden="true">/</kbd>
        {/if}
      </div>

      <div class={`absolute z-40 w-full overflow-hidden bg-background border rounded-b ${!expanded || !keyword.trim().length ? 'hidden' : ''}`}>
        <nav aria-label="검색 제안" on:mouseleave={() => (hIndex = -1)}>
          {#if pages.length}
            {#each pages as p, i (p.id)}
              <div>
                <a
                  class={`flex h-11 items-center gap-2 px-3 text-foreground hover:no-underline ${currentIndex === i ? 'focused' : ''}`}
                  href={pageUrl(p)}
                  rel="external"
                  data-sveltekit-reload
                  on:mouseenter={() => (hIndex = i)}
                  on:focus={() => (hIndex = i)}
                >
                  <ZIcon path={p.matched_title ? mdiSubdirectoryArrowRight : mdiFileDocumentOutline} size={18} class="sub flex-none text-muted-foreground" />
                  <!-- Fixed row height: rows with a second line (redirect source, description) are as tall as rows without. -->
                  <span class="flex min-w-0 flex-col justify-center">
                    <!-- eslint-disable-next-line svelte/no-at-html-tags -->
                    <span class="truncate leading-6">{@html highlight(keyword, p.title)}</span>
                    {#if p.matched_title}
                      <!-- eslint-disable-next-line svelte/no-at-html-tags -->
                      <span class="sub truncate text-xs leading-4 text-muted-foreground">{@html highlight(keyword, p.matched_title)}에서 넘어옴</span>
                    {:else if p.description}
                      <span class="sub truncate text-xs leading-4 text-muted-foreground">{p.description}</span>
                    {/if}
                  </span>
                </a>
              </div>
            {/each}
          {/if}

          <a
            class={`flex items-center gap-2 p-2 px-3 border-t text-foreground hover:no-underline ${currentIndex === pages.length ? 'focused' : ''}`}
            href={`/w/index.php?title=특수:검색&fulltext=1&search=${encodeURIComponent(keyword)}`}
            rel="external"
            data-sveltekit-reload
            on:mouseenter={() => (hIndex = pages.length)}
            on:focus={() => (hIndex = pages.length)}
          >
            <ZIcon path={mdiTextSearch} size={18} class="sub flex-none text-muted-foreground" />
            <span class="min-w-0 truncate"><b>{keyword}</b> 항목이 포함된 글 검색</span>
          </a>
        </nav>
        <div class="hidden justify-center gap-4 border-t px-3 py-1.5 text-xs text-muted-foreground md:flex" aria-hidden="true">
          <span><kbd>↑</kbd> <kbd>↓</kbd> Navigate</span>
          <span><kbd>Enter</kbd> Go</span>
          <span><kbd>Esc</kbd> Exit</span>
        </div>
      </div>
    </div>
  </div>

  <div class="flex flex-none">
    <!-- svelte-ignore a11y_accesskey -->
    <a href="/wiki/특수:임의문서" rel="external" class="navlink" title={randomTitle} accesskey="x" data-sveltekit-reload>
      <ZIcon path={mdiShuffle} class="h-5 w-5" />
      <span class="ml-1 hidden xl:inline">랜덤</span>
    </a>
    <!-- svelte-ignore a11y_accesskey -->
    <a href="/wiki/특수:최근바뀜" rel="external" class="navlink" title={recentTitle} accesskey="r" data-sveltekit-reload>
      <ZIcon path={mdiHistory} class="h-5 w-5" />
      <span class="ml-1 hidden xl:inline">바뀐글</span>
    </a>
  </div>
</div>

<style>
  .focused {
    background-color: rgb(79 70 229 / 0.7);
    color: white;
    text-decoration: none;
  }
  .focused :global(.sub) {
    color: inherit;
    opacity: 0.85;
  }
</style>
