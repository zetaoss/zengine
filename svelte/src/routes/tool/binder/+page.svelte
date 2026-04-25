<script lang="ts">
  import { mdiAlphabeticalVariant, mdiNumeric } from '@mdi/js'
  import { onMount } from 'svelte'

  import useAuthStore from '$lib/stores/auth'
  import { showToast } from '$shared/ui/toast/toast'
  import ZBadge from '$shared/ui/ZBadge.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZToggle from '$shared/ui/ZToggle.svelte'
  import httpy from '$shared/utils/httpy'
  import { displayTitle } from '$shared/utils/mediawiki'
  import { getWikiViewHref } from '$shared/utils/wikiLink'

  interface Binder {
    id: number
    title: string
    docs: number
    links: number
    title_doc: string
    enabled: boolean
    created_at: string
  }

  interface BinderUpdateResponse {
    ok: boolean
    id: number
    enabled: boolean
    deleted_id: number | null
    replacement_title: string | null
  }

  function isNewBinder(dateString: string): boolean {
    const ts = Date.parse(dateString.replace(' ', 'T') + 'Z')
    return Date.now() - ts <= 14 * 24 * 60 * 60 * 1000
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
  let updatingId: number | null = null

  const auth = useAuthStore()
  const { userInfo } = auth

  $: isSysop = ($userInfo?.groups ?? []).includes('sysop')

  function setSortMode(mode: SortMode): void {
    sortMode = mode
  }

  function getRowClass(binder: Binder): string {
    if (!binder.enabled) {
      return 'border-gray-200 opacity-75 hover:border-gray-300 dark:border-neutral-800 dark:hover:border-neutral-700'
    }

    return 'border-gray-300 hover:border-gray-400 dark:border-neutral-700 dark:hover:border-neutral-500'
  }

  function getRowAccentClass(binder: Binder): string {
    if (!binder.enabled) {
      return 'bg-(--background-color-interactive-subtle) dark:bg-(--background-color-neutral-subtle)'
    }

    return 'bg-white dark:bg-neutral-900'
  }

  function getTitleClass(binder: Binder): string {
    if (!binder.enabled) {
      return 'text-gray-500 group-hover:text-gray-600 dark:text-neutral-500 dark:group-hover:text-neutral-400'
    }

    return 'text-gray-900 group-hover:text-sky-700 dark:text-white dark:group-hover:text-sky-300'
  }

  function getBinderHref(binder: Binder): string {
    return getWikiViewHref(binder.title_doc || `Binder:${binder.title}`)
  }

  async function setBinderEnabled(binder: Binder, enabled: boolean): Promise<void> {
    if (updatingId !== null || binder.enabled === enabled) return

    updatingId = binder.id
    const [data, err] = await httpy.put<BinderUpdateResponse>(`/api/binders/${binder.id}`, { enabled })
    updatingId = null

    if (err) {
      showToast(`${enabled ? '활성화' : '비활성화'} 실패: ${err.message}`)
      return
    }

    binders = binders
      .filter((row) => row.id !== data?.deleted_id)
      .map((row) => (row.id === data?.id || row.id === binder.id ? { ...row, enabled: data?.enabled ?? enabled } : row))

    if (data?.deleted_id && data.replacement_title) {
      showToast(`바인더 넘겨줌: ${displayTitle(binder.title)} → ${displayTitle(data.replacement_title)}`)
      return
    }

    showToast(`바인더를 ${enabled ? '활성화' : '비활성화'}했어요.`)
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
      auth.update()

      const [data, err] = await httpy.get<Binder[]>('/api/binders', { _ts: Date.now() })
      if (err) {
        throw err
      }
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
                  class={`group relative block overflow-hidden rounded-2xl border no-underline shadow-sm transition duration-200 hover:no-underline hover:shadow-md ${getRowClass(binder)} ${getRowAccentClass(binder)}`}
                >
                  <div class="relative flex items-center gap-4 px-4 py-3">
                    <div class="flex min-w-0 flex-1 items-baseline gap-2">
                      <div class={`truncate text-base font-semibold transition ${getTitleClass(binder)}`}>
                        {displayTitle(binder.title)}
                      </div>
                      {#if isNewBinder(binder.created_at)}
                        <ZBadge text="N" />
                      {/if}
                    </div>

                    <div class="flex shrink-0 items-center gap-3 text-sm">
                      <span class="inline-flex items-center gap-1 tabular-nums text-gray-500 dark:text-gray-400">
                        {binder.docs}<small>/ {binder.links}</small>
                      </span>
                      {#if isSysop}
                        <!-- svelte-ignore a11y_no_static_element_interactions a11y_click_events_have_key_events -->
                        <div on:click|preventDefault|stopPropagation>
                          <ZToggle
                            checked={binder.enabled}
                            label={`${binder.title} enabled`}
                            disabled={updatingId === binder.id}
                            onchange={(event) => setBinderEnabled(binder, event.checked)}
                          />
                        </div>
                      {/if}
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
