<svelte:options runes={true} />

<script lang="ts">
  // The Instant dialog of a loop: a title typed (with search suggestions and an existence check) or picked by
  // candidateUrl, then onrun with its page ID and title. busy disables running while the parent enqueues.
  import { onDestroy, untrack } from 'svelte'

  import { showToast } from '$shared/ui/toast/toast'
  import ZModal from '$shared/ui/ZModal.svelte'
  import httpy from '$shared/utils/httpy'

  let {
    show = $bindable(false),
    heading,
    candidateUrl,
    busy,
    onrun,
  }: {
    show: boolean
    heading: string
    candidateUrl: string
    busy: boolean
    onrun: (pageID: number, title: string) => void
  } = $props()

  type CheckState = 'idle' | 'checking' | 'exists' | 'missing'

  interface SearchPage {
    id: number
    title: string
    key: string
    matched_title: string | null
    description: string | null
  }

  let fetchingCandidate = $state(false)
  let targetTitle = $state('')
  let targetPageID = $state(0)
  let checkedTitle = $state('')
  let checkState = $state<CheckState>('idle')
  let checkSeq = 0
  let autoCheckTimer: ReturnType<typeof setTimeout> | undefined
  let titleInputEl = $state<HTMLInputElement | null>(null)

  let suggestions = $state<SearchPage[]>([])
  let showSuggestions = $state(false)
  let activeSuggestionIndex = $state(-1)
  let searchAborter: AbortController | null = null
  let searchDebounceTimer: ReturnType<typeof setTimeout> | null = null

  let trimmedTitle = $derived(targetTitle.trim())

  function clearAutoCheck() {
    if (autoCheckTimer) {
      clearTimeout(autoCheckTimer)
      autoCheckTimer = undefined
    }
  }

  function resetRunModal() {
    clearAutoCheck()
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    searchAborter?.abort()
    targetTitle = ''
    targetPageID = 0
    checkedTitle = ''
    checkState = 'idle'
    suggestions = []
    showSuggestions = false
    activeSuggestionIndex = -1
    checkSeq++
  }

  // Opening starts from an empty title, focused.
  $effect(() => {
    if (show) {
      untrack(resetRunModal)
      setTimeout(() => titleInputEl?.focus(), 50)
    }
  })

  async function fetchSuggestions(q: string) {
    const trimmed = q.trim()
    if (!trimmed) {
      suggestions = []
      showSuggestions = false
      activeSuggestionIndex = -1
      return
    }

    searchAborter?.abort()
    const controller = new AbortController()
    searchAborter = controller

    try {
      const url = `/w/rest.php/v1/search/title?${new URLSearchParams({ q: trimmed, limit: '8' })}`
      const res = await fetch(url, { signal: controller.signal })
      if (!res.ok) return
      const json = (await res.json()) as { pages?: SearchPage[] }
      if (searchAborter === controller && targetTitle.trim() === trimmed) {
        suggestions = Array.isArray(json.pages) ? json.pages : []
        showSuggestions = suggestions.length > 0
        activeSuggestionIndex = -1
        const exact = suggestions.find((p) => p.title.toLowerCase() === trimmed.toLowerCase())
        if (exact) {
          checkState = 'exists'
          checkedTitle = trimmed
          targetPageID = exact.id
        }
      }
    } catch (e: unknown) {
      if (e instanceof DOMException && e.name === 'AbortError') return
    } finally {
      if (searchAborter === controller) searchAborter = null
    }
  }

  function debouncedFetchSuggestions(q: string) {
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    searchDebounceTimer = setTimeout(() => {
      void fetchSuggestions(q)
    }, 250)
  }

  function selectSuggestion(item: SearchPage) {
    clearAutoCheck()
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    searchAborter?.abort()

    targetTitle = item.title
    targetPageID = item.id
    checkedTitle = item.title
    checkState = 'exists'
    showSuggestions = false
    suggestions = []
    activeSuggestionIndex = -1
    checkSeq++
    setTimeout(() => titleInputEl?.focus(), 50)
  }

  async function checkDocumentExists(titleToCheck: string) {
    clearAutoCheck()
    const title = titleToCheck.trim()
    if (!title) {
      checkState = 'idle'
      checkedTitle = ''
      targetPageID = 0
      return
    }

    if (checkedTitle === title && checkState !== 'idle') {
      return
    }

    checkState = 'checking'
    const seq = ++checkSeq

    try {
      const [data, err] = await httpy.get<{
        query?: {
          pages?: Array<{ pageid?: number; title?: string; missing?: boolean }>
        }
      }>('/w/api.php', {
        action: 'query',
        format: 'json',
        formatversion: '2',
        titles: title,
      })
      if (seq !== checkSeq || targetTitle.trim() !== title) return
      checkedTitle = title
      const page = data?.query?.pages?.[0]
      if (!err && page && !page.missing && page.pageid && page.pageid > 0) {
        checkState = 'exists'
        targetPageID = page.pageid
        if (page.title) {
          targetTitle = page.title
          checkedTitle = page.title
        }
      } else {
        checkState = 'missing'
        targetPageID = 0
      }
    } catch {
      if (seq !== checkSeq || targetTitle.trim() !== title) return
      checkState = 'idle'
      targetPageID = 0
    }
  }

  function scheduleAutoCheck() {
    clearAutoCheck()
    const title = targetTitle.trim()
    if (!title) {
      checkState = 'idle'
      checkedTitle = ''
      return
    }
    if (checkedTitle === title && checkState !== 'idle') return

    autoCheckTimer = setTimeout(() => {
      void checkDocumentExists(title)
    }, 500)
  }

  function onTitleInput(e: Event) {
    targetTitle = (e.target as HTMLInputElement).value
    targetPageID = 0
    searchAborter?.abort()
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    checkSeq++
    checkState = 'idle'
    scheduleAutoCheck()
    debouncedFetchSuggestions(targetTitle)
  }

  function onTitleKeydown(e: KeyboardEvent) {
    if (e.isComposing) return

    if (e.key === 'ArrowDown') {
      if (showSuggestions && suggestions.length > 0) {
        e.preventDefault()
        activeSuggestionIndex = (activeSuggestionIndex + 1) % suggestions.length
        return
      }
    } else if (e.key === 'ArrowUp') {
      if (showSuggestions && suggestions.length > 0) {
        e.preventDefault()
        activeSuggestionIndex = (activeSuggestionIndex - 1 + suggestions.length) % suggestions.length
        return
      }
    } else if (e.key === 'Escape') {
      if (showSuggestions) {
        e.preventDefault()
        showSuggestions = false
        activeSuggestionIndex = -1
        return
      }
    } else if (e.key === 'Enter') {
      e.preventDefault()

      if (showSuggestions && activeSuggestionIndex >= 0 && activeSuggestionIndex < suggestions.length) {
        selectSuggestion(suggestions[activeSuggestionIndex])
        return
      }

      showSuggestions = false

      if (checkState === 'exists' && !busy) {
        handleConfirmRun()
        return
      }

      if (targetTitle.trim() && checkState !== 'checking') {
        void checkDocumentExists(targetTitle.trim())
      }
    }
  }

  async function fetchAutoCandidate() {
    clearAutoCheck()
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    searchAborter?.abort()
    showSuggestions = false
    suggestions = []
    activeSuggestionIndex = -1

    fetchingCandidate = true
    const [data, err] = await httpy.get<{ page_id: number; page_title: string }>(candidateUrl)
    fetchingCandidate = false

    if (err || !data?.page_title) {
      showToast('후보 문서를 가져오지 못했습니다.')
      return
    }

    targetTitle = data.page_title
    targetPageID = data.page_id
    checkedTitle = data.page_title
    checkSeq++
    checkState = 'exists'
    setTimeout(() => titleInputEl?.focus(), 50)
  }

  function handleConfirmRun() {
    const titleToRun = targetTitle.trim()
    if (!titleToRun || checkState !== 'exists' || busy) return
    const pageID = targetPageID
    show = false
    resetRunModal()
    onrun(pageID, titleToRun)
  }

  onDestroy(() => {
    clearAutoCheck()
    if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
    searchAborter?.abort()
  })
</script>

<ZModal
  {show}
  title={heading}
  okText="즉시 실행"
  okVariant="default"
  okDisabled={checkState !== 'exists' || busy || fetchingCandidate}
  panelClass="max-w-md !overflow-visible"
  sectionClass="p-5 !overflow-visible"
  onOk={handleConfirmRun}
  onCancel={() => {
    show = false
    resetRunModal()
  }}
>
  <div class="space-y-2 text-sm">
    <div class="flex items-center gap-2">
      <button
        type="button"
        class="shrink-0 h-9 rounded border bg-background px-3 text-xs font-medium text-foreground transition hover:bg-muted cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
        disabled={fetchingCandidate || busy}
        onclick={fetchAutoCandidate}
      >
        {#if fetchingCandidate}
          선정 중...
        {:else}
          자동선정
        {/if}
      </button>
      <div class="relative min-w-0 flex-1">
        <input
          type="text"
          bind:this={titleInputEl}
          value={targetTitle}
          class={`w-full rounded border bg-background p-2 pr-8 text-sm outline-none transition-colors ${
            checkState === 'exists'
              ? 'border-a-emerald-500 focus:border-a-emerald-600'
              : checkState === 'missing'
                ? 'border-a-rose-500 focus:border-a-rose-600'
                : checkState === 'checking'
                  ? 'border-a-blue-500 focus:border-a-blue-600'
                  : 'border-input focus:border-primary'
          }`}
          placeholder="또는 직접 문서 제목 입력"
          autocomplete="off"
          oninput={onTitleInput}
          onfocus={() => {
            if (suggestions.length > 0) showSuggestions = true
          }}
          onblur={() => {
            setTimeout(() => {
              showSuggestions = false
              activeSuggestionIndex = -1
            }, 200)
          }}
          onkeydown={onTitleKeydown}
        />
        <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2.5">
          {#if checkState === 'checking'}
            <span class="inline-block h-3.5 w-3.5 animate-spin rounded-full border-2 border-a-blue-500 border-t-transparent"></span>
          {:else if checkState === 'exists'}
            <span class="text-xs font-bold text-a-emerald-600" title="문서 확인됨">✓</span>
          {:else if checkState === 'missing'}
            <span class="text-xs font-bold text-a-rose-600" title="문서 없음">✗</span>
          {/if}
        </div>

        {#if showSuggestions && suggestions.length > 0}
          <div
            class="absolute left-0 right-0 top-full z-50 mt-1 max-h-56 overflow-y-auto rounded-md border bg-background shadow-lg text-xs"
          >
            {#each suggestions as item, i (item.id)}
              <button
                type="button"
                class={`flex w-full items-center justify-between px-3 py-2 text-left transition hover:bg-muted cursor-pointer ${
                  i === activeSuggestionIndex ? 'bg-muted text-primary font-medium' : 'text-foreground'
                }`}
                onmousedown={(e) => {
                  e.preventDefault()
                  selectSuggestion(item)
                }}
              >
                <div class="flex flex-col min-w-0">
                  <span class="truncate font-medium">{item.title}</span>
                  {#if item.matched_title}
                    <span class="text-[11px] text-muted-foreground truncate">{item.matched_title}에서 넘어옴</span>
                  {:else if item.description}
                    <span class="text-[11px] text-muted-foreground truncate">{item.description}</span>
                  {/if}
                </div>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    </div>
    {#if checkState === 'missing'}
      <p class="text-xs text-a-rose-500">
        '{trimmedTitle}' 문서를 찾을 수 없습니다.
      </p>
    {/if}
  </div>
</ZModal>
