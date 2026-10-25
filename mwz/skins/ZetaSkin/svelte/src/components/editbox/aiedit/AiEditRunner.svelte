<svelte:options runes={true} />

<script lang="ts">
  import { mdiClose, mdiDotsVertical, mdiPlay, mdiPlus } from '@mdi/js'
  import type { Snippet } from 'svelte'

  import CButton from '$shared/ui/CButton.svelte'
  import CMenu from '$shared/ui/CMenu.svelte'
  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZModal from '$shared/ui/ZModal.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import ZTextarea from '$shared/ui/ZTextarea.svelte'
  import httpy from '$shared/utils/httpy'

  import AiEditTaskResults from './AiEditTaskResults.svelte'
  import type {
    AIEditPromptForRunner,
    AIEditRenderedMessage,
    AIEditRequestPayload,
    AIEditRequestType,
    AIEditStoreResult,
    AIEditTaskPhase,
    AIEditUserInput,
    AIEditUserInputValue,
  } from './aiEditTypes'

  let {
    prompt,
    title,
    pageId,
    editorContent,
    userInputs,
    submit,
    preview,
    promptHeader = undefined,
    onPromptListChanged = undefined,
    onRequestTypeChange = undefined,
  }: {
    prompt: AIEditPromptForRunner
    title: string
    pageId: number | undefined
    editorContent: string
    userInputs: AIEditUserInput[]
    submit: (payload: AIEditRequestPayload) => Promise<AIEditStoreResult>
    preview: (payload: AIEditRequestPayload) => Promise<AIEditRenderedMessage[]>
    promptHeader?: Snippet
    onPromptListChanged?: () => void | Promise<void>
    onRequestTypeChange?: (newType: AIEditRequestType) => void
  } = $props()

  let submitting = $state(false)
  let submitError = $state('')
  let polling = $state(false)
  let submittedTaskPhase = $state<AIEditTaskPhase | undefined>(undefined)
  let submittedTaskId = $state<number | undefined>(undefined)
  let taskResetToken = $state(0)
  let values = $state<Record<string, AIEditUserInputValue>>({})
  let previewModalOpen = $state(false)
  let previewLoading = $state(false)
  let previewMessages = $state<AIEditRenderedMessage[]>([])
  let saveAsModalOpen = $state(false)
  let saveAsTitle = $state('')
  let savingPromptAs = $state(false)

  let hasPromptContent = $derived(prompt.content.trim().length > 0)
  let trimmedTitle = $derived(title.trim())
  let canSubmit = $derived(hasPromptContent && trimmedTitle.length > 0 && !submitting)
  let displayedTaskStatus = $derived(submitting ? 'Creating' : submittedTaskPhase === 'Completed' ? undefined : submittedTaskPhase)

  function textValue(name: string) {
    const v = values[name]
    return typeof v === 'string' ? v : ''
  }

  function textsValue(name: string) {
    const v = values[name]
    return Array.isArray(v) ? v : []
  }

  function setText(name: string, value: string) {
    values = { ...values, [name]: value }
  }

  function setTexts(name: string, items: string[]) {
    values = { ...values, [name]: items }
  }

  function payload(): AIEditRequestPayload {
    const userValues: Record<string, AIEditUserInputValue> = {}
    for (const input of userInputs) {
      if (values[input.name] !== undefined) userValues[input.name] = values[input.name]
    }
    return {
      pageId,
      title: trimmedTitle,
      promptId: prompt.id,
      promptTitle: prompt.title,
      requestType: prompt.requestType,
      wikitext: editorContent,
      userInputs: userValues,
    }
  }

  async function handleSubmitClick() {
    if (!canSubmit) return

    submitting = true
    submitError = ''
    resetResultState()

    try {
      const data = await submit(payload())
      submittedTaskId = data.id
    } catch (error) {
      submitError = error instanceof Error && error.message ? error.message : 'AI 편집 등록에 실패했습니다.'
    } finally {
      submitting = false
    }
  }

  async function openPreview() {
    previewModalOpen = true
    previewLoading = true
    previewMessages = []
    try {
      previewMessages = await preview(payload())
    } catch (error) {
      previewModalOpen = false
      showToast(error instanceof Error && error.message ? error.message : '미리보기를 만들지 못했습니다.')
    } finally {
      previewLoading = false
    }
  }

  function closePreview() {
    previewModalOpen = false
  }

  function resetResultState() {
    polling = false
    submittedTaskPhase = undefined
    submittedTaskId = undefined
    taskResetToken += 1
  }

  function openSaveAsModal() {
    saveAsTitle = `${prompt.title} 복사본`
    saveAsModalOpen = true
  }

  async function savePromptAs() {
    const newTitle = saveAsTitle.trim()
    if (!newTitle) return

    savingPromptAs = true
    try {
      const [, err] = await httpy.post('/api/ai-prompts', {
        title: newTitle,
        request_type: prompt.requestType,
        content: prompt.content,
        model: prompt.model,
      })
      if (err) {
        showToast(err.message || '새 프롬프트 저장에 실패했습니다.')
        return
      }
      saveAsModalOpen = false
      showToast('새 프롬프트로 저장했습니다.')
      void onPromptListChanged?.()
    } catch (error) {
      showToast(error instanceof Error && error.message ? error.message : '새 프롬프트 저장에 실패했습니다.')
    } finally {
      savingPromptAs = false
    }
  }
</script>

<div class="flex h-full min-h-0 flex-col gap-3">
  <AiEditTaskResults
    {pageId}
    {submittedTaskId}
    resetToken={taskResetToken}
    requestType={prompt.requestType}
    title={trimmedTitle}
    onPollingChange={(value) => (polling = value)}
    onPhaseChange={(phase) => (submittedTaskPhase = phase)}
    {onRequestTypeChange}
  />

  <div class="flex min-h-0 flex-1 flex-col gap-3 border bg-a-gray-100 p-4">
    {@render promptHeader?.()}
    {#if !hasPromptContent}
      <div class="rounded border border-a-amber-200 bg-a-amber-50 px-3 py-2 text-sm text-a-amber-700">프롬프트 내용이 없습니다.</div>
    {/if}

    <div class="relative">
      <pre
        id="ai-edit-runner-prompt"
        class="max-h-60 overflow-auto whitespace-pre-wrap rounded border bg-a-gray-50 p-3 font-mono text-xs leading-relaxed">{prompt.content}</pre>
      <div class="absolute right-5 top-1 flex items-center">
        {#if displayedTaskStatus}
          <div class="flex items-center rounded-lg bg-a-gray-500/20 px-2 py-1 text-xs text-a-slate-500">
            {#if submitting || polling}
              <ZSpinner size="0.875rem" extraClass="mr-1" />
            {/if}
            {displayedTaskStatus}
          </div>
        {/if}
        <CButton type="button" variant="subtle" size="small" disabled={!canSubmit || polling} onclick={() => void handleSubmitClick()}>
          <ZIcon path={mdiPlay} />
          Run
        </CButton>
        <CMenu>
          {#snippet trigger({ toggle })}
            <CButton type="button" variant="subtle" size="icon-sm" aria-label="프롬프트 메뉴" onclick={toggle}>
              <ZIcon path={mdiDotsVertical} />
            </CButton>
          {/snippet}
          {#snippet menu({ close })}
            <div class="flex flex-col items-stretch">
              <CButton
                type="button"
                variant="ghost"
                class="w-full justify-start! rounded-none!"
                disabled={!canSubmit}
                onclick={() => {
                  void openPreview()
                  close()
                }}
              >
                Render
              </CButton>
              <CButton
                type="button"
                variant="ghost"
                class="w-full justify-start! rounded-none!"
                disabled={!hasPromptContent || savingPromptAs}
                onclick={() => {
                  openSaveAsModal()
                  close()
                }}
              >
                Save As...
              </CButton>
            </div>
          {/snippet}
        </CMenu>
      </div>
    </div>

    {#if userInputs.length > 0}
      <div class="flex min-h-0 flex-col gap-3 overflow-auto text-sm">
        {#each userInputs as input (input.name)}
          {@const label = input.label ?? input.name}
          {#if input.type === 'bool'}
            <label class="flex items-center gap-2">
              <input
                type="checkbox"
                checked={values[input.name] === true}
                onchange={(e) => (values = { ...values, [input.name]: e.currentTarget.checked })}
              />
              {label}
            </label>
          {:else if input.type === 'texts'}
            <div class="flex flex-col gap-2">
              {#each textsValue(input.name) as item, i (i)}
                <div class="flex items-start gap-1">
                  <div class="min-w-0 flex-1 text-xs">
                    <ZTextarea
                      id={`ai-edit-input-${input.name}-${i}`}
                      modelValue={item}
                      placeholder={`${label} ${i + 1}`}
                      maxHeight={240}
                      onUpdateModelValue={(v: string) => setTexts(input.name, textsValue(input.name).map((x, j) => (j === i ? v : x)))}
                    />
                  </div>
                  <CButton
                    type="button"
                    variant="ghost"
                    size="icon-sm"
                    aria-label={`${label} ${i + 1} 빼기`}
                    onclick={() => setTexts(input.name, textsValue(input.name).filter((_, j) => j !== i))}
                  >
                    <ZIcon path={mdiClose} />
                  </CButton>
                </div>
              {/each}
              <div>
                <CButton type="button" variant="outline" size="small" onclick={() => setTexts(input.name, [...textsValue(input.name), ''])}>
                  <ZIcon path={mdiPlus} />
                  {label}
                </CButton>
              </div>
            </div>
          {:else}
            <div class="flex flex-col gap-1 text-xs">
              <span class="font-semibold">{label}</span>
              <ZTextarea
                id={`ai-edit-input-${input.name}`}
                modelValue={textValue(input.name)}
                placeholder={undefined}
                maxHeight={160}
                onUpdateModelValue={(v: string) => setText(input.name, v)}
              />
            </div>
          {/if}
        {/each}
      </div>
    {/if}

    <div class="flex flex-col items-start gap-2">
      {#if submitError}
        <div class="text-sm text-a-red-500">{submitError}</div>
      {/if}
    </div>
  </div>
</div>

<ZModal
  show={previewModalOpen}
  title="Rendered Prompt"
  okText="닫기"
  okVariant="default"
  showCancel={false}
  panelClass="max-w-[calc(100vw-2rem)] md:max-w-4xl"
  sectionClass="min-h-0 flex-1 overflow-hidden p-5"
  onOk={closePreview}
  onCancel={closePreview}
>
  <div class="flex max-h-[calc(100dvh-14rem)] flex-col gap-3 overflow-auto">
    {#if previewLoading}
      <ZSpinner />
    {/if}
    {#each previewMessages as m, i (i)}
      <div class="rounded border border-a-slate-200">
        <div class="border-b border-a-slate-200 bg-a-slate-100 px-3 py-1 font-mono text-xs text-a-slate-500">{m.role}</div>
        <pre class="whitespace-pre-wrap bg-a-slate-50 p-3 font-mono text-xs text-a-slate-800">{m.content}</pre>
      </div>
    {/each}
  </div>
</ZModal>

<ZModal
  show={saveAsModalOpen}
  title="새 프롬프트로 저장"
  okText={savingPromptAs ? '저장 중...' : '저장'}
  okVariant="default"
  okDisabled={!saveAsTitle.trim() || savingPromptAs}
  closable={!savingPromptAs}
  backdropClosable={!savingPromptAs}
  onOk={savePromptAs}
  onCancel={() => {
    if (!savingPromptAs) saveAsModalOpen = false
  }}
>
  <label class="flex flex-col gap-2 text-sm text-a-slate-700">
    <span class="font-semibold">프롬프트 제목</span>
    <input class="border-border rounded p-1 bg-background" bind:value={saveAsTitle} placeholder="새 프롬프트 제목" />
  </label>
</ZModal>
