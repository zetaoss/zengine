<svelte:options customElement={{ tag: 'c-search', shadow: 'none' }} />

<script lang="ts">
  import { mdiHistory, mdiMagnify, mdiShuffle } from '@mdi/js'
  import { onDestroy } from 'svelte'

  import { useDismissable } from '$shared/composables/useDismissable'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import getShortcut from '$shared/utils/shortcut'

  interface Page {
    description: string
    excerpt: string
    id: number
    key: string
    matched_title: string
    thumbnail: Record<string, unknown>
    title: string
  }
  interface SearchResponse {
    pages?: Page[]
    [k: string]: unknown
  }

  let root: HTMLElement | null = null
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
    if (keyword.trim()) expanded = true
  }

  function handleUpDown(offset: number) {
    const max = pages.length
    let next = kIndex + offset
    if (next < -1) next = max
    if (next > max) next = -1
    kIndex = next
  }

  function goToSearch() {
    if (typeof window === 'undefined') return

    const base = new URL('/w/index.php', window.location.origin)
    base.searchParams.set('title', '특수:검색')

    if (kIndex < 0) {
      base.searchParams.set('search', displayQuery)
    } else if (kIndex === pages.length) {
      base.searchParams.set('search', displayQuery)
      base.searchParams.set('fulltext', '1')
      base.searchParams.set('ns0', '1')
    } else {
      base.searchParams.set('search', pages[kIndex]?.title ?? displayQuery)
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
      close()
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

<div class="z-text ml-auto flex w-full md:max-w-2xl">
  <div class="m-1.5 grow">
    <div bind:this={root} class="relative" role="search">
      <div class={`flex h-9 z-base rounded-t ${!expanded || !keyword.trim().length ? 'rounded-b' : ''}`}>
        <!-- svelte-ignore a11y_accesskey -->
        <input
          accesskey="f"
          aria-label="search"
          type="search"
          class="grow h-full bg-transparent px-3 outline-0"
          name="search"
          placeholder="검색..."
          title={searchTitle}
          autocomplete="off"
          value={displayQuery}
          on:input={onInput}
          on:focus={onFocus}
          on:keydown={onKeyDown}
        />
        <button type="button" class="z-10 h-full w-12 flex-none bg-transparent focus:text-blue-700" on:click={onClick}>
          <ZIcon path={mdiMagnify} size={24} />
        </button>
      </div>

      <div class={`absolute z-40 w-full z-base border rounded-b ${!expanded || !keyword.trim().length ? 'hidden' : ''}`}>
        <nav aria-label="검색 제안" on:mouseleave={() => (hIndex = -1)}>
          {#if pages.length}
            {#each pages as p, i (p.id)}
              <div>
                <a
                  class={`block p-1.5 px-3 z-text ${currentIndex === i ? 'focused' : ''}`}
                  href={`/w/index.php?title=특수:검색&search=${encodeURIComponent(p.title)}`}
                  rel="external"
                  data-sveltekit-reload
                  on:mouseenter={() => (hIndex = i)}
                  on:focus={() => (hIndex = i)}
                >
                  <!-- eslint-disable-next-line svelte/no-at-html-tags -->
                  <span>{@html highlight(keyword, p.title)}</span>
                </a>
              </div>
            {/each}
          {/if}

          <a
            class={`block p-2 px-3 border-t rounded-b z-text ${currentIndex === pages.length ? 'focused' : ''}`}
            href={`/w/index.php?title=특수:검색&fulltext=1&search=${encodeURIComponent(keyword)}`}
            rel="external"
            data-sveltekit-reload
            on:mouseenter={() => (hIndex = pages.length)}
            on:focus={() => (hIndex = pages.length)}
          >
            <ZIcon path={mdiMagnify} />
            <b>{keyword}</b> 항목이 포함된 글 검색
          </a>
        </nav>
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
    color: #fff;
    text-decoration: none;
  }
</style>
