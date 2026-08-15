<svelte:options runes={true} />

<script lang="ts">
  import { showToast } from '$shared/ui/toast/toast'
  import ZSelect from '$shared/ui/ZSelect.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import getRLCONF from '$lib/utils/rlconf'
  import httpy from '$shared/utils/httpy'

  import { subscribeWikiEditorContent } from '../wikiEditor'
  import AiEditRunner from './AiEditRunner.svelte'
  import type { AIEditPromptForRunner, AIEditRequestType, AIEditRunnerSubmitPayload, AIEditStoreResult } from './aiEditTypes'

  interface PromptItem {
    id: number
    user_id: number
    title: string
    content: string
    request_type: string
    use_count: number
    user_name: string
    created_at: string
  }

  interface PromptSelectItem {
    value: string
    label: string
  }

  let {
    requestType,
    pageId,
    title,
  }: {
    requestType: AIEditRequestType
    pageId: number | undefined
    title: string
  } = $props()

  let currentRequestType = $derived(requestType)

  function handleRequestTypeChange(newType: AIEditRequestType) {
    currentRequestType = newType
    selectPreferredPrompt(newType)
  }

  let promptItems = $state<PromptItem[]>([])
  let promptTitle = $state('')
  let promptListLoading = $state(false)
  let mwEditorContent = $state('')

  const promptStorageKeyPrefix = 'ai-edit-prompt:'

  $effect(() => {
    return subscribeWikiEditorContent((content) => {
      mwEditorContent = content
    })
  })

  let sortedPromptItems = $derived.by(() => {
    const createPrompts = promptItems.filter((p) => p.request_type === 'create').sort((a, b) => b.use_count - a.use_count)
    const editPrompts = promptItems.filter((p) => p.request_type === 'edit').sort((a, b) => b.use_count - a.use_count)

    if (currentRequestType === 'create') {
      return [...createPrompts, ...editPrompts]
    } else {
      return [...editPrompts, ...createPrompts]
    }
  })
  let currentPromptItem = $derived(promptItems.find((p) => p.title === promptTitle))
  let currentRunnerPrompt = $derived<AIEditPromptForRunner | null>(
    currentPromptItem && isAIEditRequestType(currentPromptItem.request_type)
      ? {
          id: currentPromptItem.id,
          userId: currentPromptItem.user_id,
          title: currentPromptItem.title,
          requestType: currentPromptItem.request_type,
          content: currentPromptItem.content,
        }
      : null,
  )
  let promptSelectItems = $derived<PromptSelectItem[]>(
    sortedPromptItems.length > 0
      ? sortedPromptItems.map((p) => ({
          group: p.request_type === 'create' ? '생성용' : '편집용',
          value: p.title,
          label: p.title,
        }))
      : [{ value: '', label: '프롬프트 없음' }],
  )

  $effect(() => {
    void fetchPromptList()
  })

  $effect(() => {
    if (!sortedPromptItems.length) {
      promptTitle = ''
      return
    }

    if (!sortedPromptItems.some((p) => p.title === promptTitle)) {
      selectPreferredPrompt(currentRequestType)
    }
  })

  function isAIEditRequestType(value: string): value is AIEditRequestType {
    return value === 'create' || value === 'edit'
  }

  function getStoredPromptTitle(type: AIEditRequestType) {
    try {
      return window.localStorage.getItem(`${promptStorageKeyPrefix}${type}`) ?? ''
    } catch {
      return ''
    }
  }

  function selectPreferredPrompt(type: AIEditRequestType) {
    const promptsForType = promptItems.filter((prompt) => prompt.request_type === type).sort((a, b) => b.use_count - a.use_count)
    const storedPromptTitle = getStoredPromptTitle(type)
    const preferredPrompt = promptsForType.find((prompt) => prompt.title === storedPromptTitle) ?? promptsForType[0] ?? sortedPromptItems[0]
    promptTitle = preferredPrompt?.title ?? ''
  }

  function storePromptTitle(value: string) {
    promptTitle = value
    try {
      window.localStorage.setItem(`${promptStorageKeyPrefix}${currentRequestType}`, value)
    } catch {
      // Continue with the in-memory selection when storage is unavailable.
    }
  }

  function formatPromptDate(value: string) {
    const datePart = value.match(/^\d{4}-\d{2}-\d{2}/)?.[0]
    if (datePart) return datePart

    const date = new Date(value)
    return Number.isNaN(date.getTime()) ? '' : date.toISOString().slice(0, 10)
  }

  async function fetchPromptList() {
    promptListLoading = true
    try {
      const [data, err] = await httpy.get<PromptItem[]>('/api/ai-prompts')
      if (err) {
        console.error(err)
        showToast(err.message || '프롬프트 목록을 불러오지 못했습니다.')
        return
      }
      promptItems = data ?? []
    } catch (error) {
      console.error(error)
      showToast('프롬프트 목록을 불러오지 못했습니다.')
    } finally {
      promptListLoading = false
    }
  }

  async function refreshPromptList() {
    await fetchPromptList()
  }

  async function submitRunner(payload: AIEditRunnerSubmitPayload) {
    const [data, err] = await httpy.post<AIEditStoreResult>('/api/ai-edit', {
      page_id: payload.pageId ?? pageId,
      title: payload.title,
      prompt_title: payload.promptTitle,
      request_type: payload.requestType,
      llm_input: payload.llmInput,
    })

    if (err) throw err
    if (!data) throw new Error('AI 편집 등록에 실패했습니다.')
    return data
  }
</script>

{#snippet promptHeader()}
  <div class="flex items-center gap-2">
    <a href="/tool/ai-prompts" target="_blank" rel="noopener noreferrer" class="external text-xs" title="프롬프트 관리 새 창에서 열기">
      프롬프트
    </a>
    {#if promptListLoading}
      <ZSpinner />
    {:else}
      <ZSelect bind:value={promptTitle} items={promptSelectItems} class="min-w-0 flex-1" onchange={storePromptTitle}>
        {#snippet item(p: PromptSelectItem)}
          {@const promptMeta = promptItems.find((item) => item.title === p.value)}
          {@const promptAuthor = promptMeta && promptMeta.user_id !== getRLCONF()?.wgUserId ? promptMeta.user_name : ''}
          <div class="grid flex-1 grid-cols-[minmax(0,1fr)_auto_auto_auto] items-center gap-3 overflow-hidden">
            <div class="truncate">{p.label}</div>
            <span class="truncate text-right text-xs opacity-50">{promptAuthor}</span>
            <span class="text-right text-xs tabular-nums opacity-50">{promptMeta?.use_count ?? 0} runs</span>
            <span class="text-right text-xs tabular-nums opacity-50">{promptMeta ? formatPromptDate(promptMeta.created_at) : ''}</span>
          </div>
        {/snippet}
      </ZSelect>
    {/if}
  </div>
{/snippet}

{#if currentRunnerPrompt}
  <AiEditRunner
    prompt={currentRunnerPrompt}
    {title}
    {pageId}
    editorContent={mwEditorContent}
    submit={submitRunner}
    {promptHeader}
    onPromptListChanged={refreshPromptList}
    onRequestTypeChange={handleRequestTypeChange}
  />
{:else if !promptListLoading}
  <div class="flex h-full min-h-0 flex-col gap-3 border bg-a-gray-100 p-4">
    {@render promptHeader()}
    <div class="flex min-h-0 flex-1 items-center justify-center rounded border border-dashed border-a-slate-300 text-sm text-a-slate-500">
      사용할 수 있는 프롬프트가 없습니다.
    </div>
  </div>
{/if}
