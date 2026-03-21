<script lang="ts">
  import { mdiAlphabeticalVariant, mdiNumeric } from '@mdi/js'
  import { onMount } from 'svelte'

  import ZIcon from '$shared/ui/ZIcon.svelte'
  import { displayTitle } from '$shared/utils/mediawiki'
  import { getWikiViewHref } from '$shared/utils/wikiLink'

  interface Binder {
    id: number
    title: string
    docs: number
    links: number
    title_doc: string
  }

  type SortMode = 'docs' | 'title'
  type BinderSection = {
    key: 'hundred_plus' | 'fifty_plus' | 'twenty_five_plus' | 'rest'
    title: string
    binders: Binder[]
  }

  let binders: Binder[] = []
  let loading = true
  let error: string | null = null
  let sortMode: SortMode = 'title'

  function setSortMode(mode: SortMode): void {
    sortMode = mode
  }

  function getRowClass(): string {
    return 'border-gray-300 hover:border-gray-400 dark:border-neutral-700 dark:hover:border-neutral-500'
  }

  function getRowAccentClass(): string {
    return 'bg-white dark:bg-neutral-900'
  }

  function getBinderHref(binder: Binder): string {
    return getWikiViewHref(binder.title_doc || `Binder:${binder.title}`)
  }

  function compareBinders(a: Binder, b: Binder, mode: SortMode): number {
    if (mode === 'title') {
      const byTitle = displayTitle(a.title).localeCompare(displayTitle(b.title), 'ko')
      return byTitle || b.docs - a.docs || a.id - b.id
    }

    return b.docs - a.docs || a.id - b.id
  }

  $: sortedBinders = [...binders].sort((a, b) => compareBinders(a, b, sortMode))
  $: sections = [
    {
      key: 'hundred_plus',
      title: '100+',
      binders: sortedBinders.filter((binder) => binder.docs >= 100),
    },
    {
      key: 'fifty_plus',
      title: '50+',
      binders: sortedBinders.filter((binder) => binder.docs >= 50 && binder.docs < 100),
    },
    {
      key: 'twenty_five_plus',
      title: '25+',
      binders: sortedBinders.filter((binder) => binder.docs >= 25 && binder.docs < 50),
    },
    {
      key: 'rest',
      title: '나머지',
      binders: sortedBinders.filter((binder) => binder.docs < 25),
    },
  ] satisfies BinderSection[]

  onMount(async () => {
    try {
      const response = await fetch('/w/rest.php/binder')
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }
      const data = await response.json()
      binders = data || []
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load binders'
    } finally {
      loading = false
    }
  })
</script>

<div class="p-6">
  <div class="mb-5 flex flex-wrap items-end justify-between gap-4">
    <div class="flex items-baseline gap-2">
      <h1 class="text-2xl font-bold tracking-tight">바인더</h1>
      {#if !loading && !error}
        <p class="text-sm text-gray-500 dark:text-gray-400">{binders.length}</p>
      {/if}
    </div>

    <div
      class="inline-flex overflow-hidden rounded-xl border border-gray-300 bg-white shadow-sm dark:border-neutral-700 dark:bg-neutral-900"
    >
      <button
        type="button"
        class={`grid h-10 w-10 cursor-pointer place-items-center border-r border-gray-300 transition dark:border-neutral-700 ${
          sortMode === 'title'
            ? 'bg-gray-200 text-black dark:bg-neutral-800 dark:text-white'
            : 'text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-neutral-800'
        }`}
        title="Title"
        aria-label="Title"
        aria-pressed={sortMode === 'title'}
        on:click={() => setSortMode('title')}
      >
        <ZIcon path={mdiAlphabeticalVariant} size={18} />
      </button>
      <button
        type="button"
        class={`grid h-10 w-10 cursor-pointer place-items-center transition ${
          sortMode === 'docs'
            ? 'bg-gray-200 text-black dark:bg-neutral-800 dark:text-white'
            : 'text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-neutral-800'
        }`}
        title="Docs"
        aria-label="Docs"
        aria-pressed={sortMode === 'docs'}
        on:click={() => setSortMode('docs')}
      >
        <ZIcon path={mdiNumeric} size={18} />
      </button>
    </div>
  </div>

  {#if loading}
    <p>로딩 중...</p>
  {:else if error}
    <p class="text-red-500">오류: {error}</p>
  {:else}
    <div class="space-y-8">
      {#each sections as section (section.key)}
        {#if section.binders.length > 0}
          <section>
            <div class="mb-3 flex items-baseline gap-2">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{section.title}</h2>
              <p class="text-sm text-gray-500 dark:text-gray-400">{section.binders.length}</p>
            </div>

            <div class="grid gap-3 lg:grid-cols-2 2xl:grid-cols-3">
              {#each section.binders as binder (binder.id)}
                <a
                  href={getBinderHref(binder)}
                  rel="external"
                  data-sveltekit-reload
                  class={`group relative block overflow-hidden rounded-2xl border no-underline shadow-sm transition duration-200 hover:no-underline hover:shadow-md ${getRowClass()} ${getRowAccentClass()}`}
                >
                  <div class="relative flex items-center gap-4 px-4 py-3">
                    <div class="min-w-0 flex-1">
                      <div
                        class="truncate text-base font-semibold text-gray-900 transition group-hover:text-sky-700 dark:text-white dark:group-hover:text-sky-300"
                      >
                        {displayTitle(binder.title)}
                      </div>
                    </div>

                    <div class="flex shrink-0 items-center gap-3 text-sm">
                      <span class="tabular-nums text-gray-500 dark:text-gray-400">{binder.docs} / {binder.links}</span>
                    </div>
                  </div>
                </a>
              {/each}
            </div>
          </section>
        {/if}
      {/each}
    </div>
  {/if}
</div>
