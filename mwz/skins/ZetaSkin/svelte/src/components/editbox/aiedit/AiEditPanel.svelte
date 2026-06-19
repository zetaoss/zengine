<svelte:options runes={true} />

<script lang="ts">
  import { mdiEye, mdiRefresh } from '@mdi/js'

  import AiEditRunner from '$shared/components/ai-edit/AiEditRunner.svelte'
  import type { AIEditPromptForRunner, AIEditRequestType, AIEditRunnerSubmitPayload } from '$shared/types/aiEdit'
  import CButton from '$shared/ui/CButton.svelte'
  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSelect from '$shared/ui/ZSelect.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import httpy from '$shared/utils/httpy'

  interface PromptItem {
    id: number
    title: string
    content: string
    request_type: string
    use_count: number
    is_favorite: boolean
  }

  interface PromptSelectItem {
    value: string
    label: string
  }

  interface StoreResp {
    ok: boolean
    id: number
    created: boolean
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

  let promptItems = $state<PromptItem[]>([])
  let promptTitle = $state('')
  let promptListLoading = $state(false)
  let mwEditorContent = $state('')
  let mwEditorContentDirty = $state(false)
  let contentVersion = $state(0)

  $effect(() => {
    const textarea = document.getElementById('wpTextbox1') as HTMLTextAreaElement | null
    if (!textarea) return

    mwEditorContent = textarea.value
    mwEditorContentDirty = false

    const handleInput = () => {
      mwEditorContentDirty = textarea.value !== mwEditorContent
    }

    textarea.addEventListener('input', handleInput)
    return () => textarea.removeEventListener('input', handleInput)
  })

  let filteredPromptItems = $derived(
    promptItems
      .filter((p) => p.request_type === requestType)
      .sort((a, b) => {
        if (a.is_favorite !== b.is_favorite) return a.is_favorite ? -1 : 1
        return b.use_count - a.use_count
      }),
  )
  let currentPromptItem = $derived(filteredPromptItems.find((p) => p.title === promptTitle))
  let currentRunnerPrompt = $derived<AIEditPromptForRunner | null>(
    currentPromptItem && isAIEditRequestType(currentPromptItem.request_type)
      ? {
          id: currentPromptItem.id,
          title: currentPromptItem.title,
          requestType: currentPromptItem.request_type,
          content: currentPromptItem.content,
        }
      : null,
  )
  let promptSelectItems = $derived<PromptSelectItem[]>(
    filteredPromptItems.length > 0
      ? filteredPromptItems.map((p) => ({ value: p.title, label: p.title }))
      : [{ value: '', label: '프롬프트 없음' }],
  )

  $effect(() => {
    void fetchPromptList()
  })

  $effect(() => {
    if (!filteredPromptItems.length) {
      promptTitle = ''
      return
    }
    if (!filteredPromptItems.some((p) => p.title === promptTitle)) {
      promptTitle = filteredPromptItems[0].title
    }
  })

  function isAIEditRequestType(value: string): value is AIEditRequestType {
    return value === 'create' || value === 'edit'
  }

  async function fetchPromptList() {
    promptListLoading = true
    try {
      const [data, err] = await httpy.get<PromptItem[]>('/api/ai-edit/prompts')
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

  function releaseEditWarning() {
    window.jQuery?.(window).off('beforeunload.editwarning')
  }

  function resetContentFromEditor() {
    const textarea = document.getElementById('wpTextbox1') as HTMLTextAreaElement | null
    mwEditorContent = textarea?.value ?? ''
    mwEditorContentDirty = false
    contentVersion += 1
  }

  async function submitRunner(payload: AIEditRunnerSubmitPayload) {
    const [data, err] = await httpy.post<StoreResp>('/api/ai-edit', {
      page_id: payload.pageId ?? pageId,
      title: payload.title,
      prompt_title: payload.promptTitle,
      request_type: payload.requestType,
      llm_input: payload.llmInput,
    })

    if (err) throw err
    if (!data) throw new Error('AI 편집 등록에 실패했습니다.')

    releaseEditWarning()
    window.location.assign(`/tool/ai-edit/tasks/${data.id}`)
  }
</script>

<div class="flex h-full min-h-0 flex-col gap-3 border p-4">
  <div class="flex items-center gap-2">
    <span class="w-20 shrink-0 text-right text-sm font-semibold text-a-slate-700">템플릿</span>
    {#if promptListLoading}
      <ZSpinner />
    {:else}
      <ZSelect bind:value={promptTitle} items={promptSelectItems} class="min-w-0 flex-1">
        {#snippet item(p: PromptSelectItem)}
          {@const promptMeta = filteredPromptItems.find((item) => item.title === p.value)}
          <div class="flex flex-1 items-center justify-between gap-2 overflow-hidden">
            <div class="truncate">{p.label}</div>
            <div class="w-20 shrink-0 text-right text-xs opacity-50">
              {#if promptMeta?.is_favorite}⭐
              {/if}{promptMeta?.use_count ?? 0} runs
            </div>
          </div>
        {/snippet}
      </ZSelect>
      <div class="flex shrink-0 items-center gap-1">
        {#if currentPromptItem}
          <CButton
            href={`/tool/ai-edit/prompts/${currentPromptItem.id}`}
            target="_blank"
            title={`${promptTitle} 편집`}
            size="small"
            variant="outline"
          >
            <ZIcon path={mdiEye} />
          </CButton>
        {/if}
        <CButton type="button" variant="outline" size="small" title="새로고침" onclick={() => void refreshPromptList()}>
          <ZIcon path={mdiRefresh} />
        </CButton>
      </div>
    {/if}
  </div>

  {#if currentRunnerPrompt}
    <AiEditRunner
      prompt={currentRunnerPrompt}
      initialTitle={title}
      initialPageId={pageId}
      externalContent={mwEditorContent}
      externalContentDirty={mwEditorContentDirty}
      externalContentVersion={contentVersion}
      onContentResetRequest={resetContentFromEditor}
      submit={submitRunner}
      layout="compact"
      hideTitle={true}
      hideReservedStatus={true}
      class="border-0! p-0!"
    />
  {:else if !promptListLoading}
    <div class="flex min-h-0 flex-1 items-center justify-center rounded border border-dashed border-a-slate-300 text-sm text-a-slate-500">
      사용할 수 있는 프롬프트가 없습니다.
    </div>
  {/if}
</div>
