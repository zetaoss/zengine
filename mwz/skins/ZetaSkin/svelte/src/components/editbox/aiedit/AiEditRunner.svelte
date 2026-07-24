<svelte:options runes={true} />

<script lang="ts">
  import { mdiDotsVertical, mdiPlay } from '@mdi/js'
  import { untrack } from 'svelte'

  import getRLCONF from '$lib/utils/rlconf'
  import CButton from '$shared/ui/CButton.svelte'
  import CMenu from '$shared/ui/CMenu.svelte'
  import { showConfirm } from '$shared/ui/confirm/confirm'
  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZModal from '$shared/ui/ZModal.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import ZTextarea from '$shared/ui/ZTextarea.svelte'
  import httpy from '$shared/utils/httpy'

  import AiEditTaskResults from './AiEditTaskResults.svelte'
  import type {
    AIEditPromptForRunner,
    AIEditRequestType,
    AIEditRunnerSubmitPayload,
    AIEditStoreResult,
    AIEditTaskPhase,
  } from './aiEditTypes'
  import { renderFinalPrompt } from './prompt/promptRenderer'

  type SubmitHandler = (payload: AIEditRunnerSubmitPayload) => Promise<AIEditStoreResult>

  let {
    prompt,
    title,
    pageId,
    editorContent,
    submit,
    onPromptListChanged = undefined,
    onRequestTypeChange = undefined,
  }: {
    prompt: AIEditPromptForRunner
    title: string
    pageId: number | undefined
    editorContent: string
    submit: SubmitHandler
    onPromptListChanged?: () => void | Promise<void>
    onRequestTypeChange?: (newType: AIEditRequestType) => void
  } = $props()

  let submitting = $state(false)
  let submitError = $state('')
  let polling = $state(false)
  let submittedTaskPhase = $state<AIEditTaskPhase | undefined>(undefined)
  let submittedTaskId = $state<number | undefined>(undefined)
  let taskResetToken = $state(0)
  let workingTemplate = $state('')
  let lastRunnerIdentity = $state('')
  let lastRunnerContextIdentity = $state('')
  let lastSavedContent: string | undefined
  let previewModalOpen = $state(false)
  let savingPrompt = $state(false)
  let saveAsModalOpen = $state(false)
  let saveAsTitle = $state('')
  let savingPromptAs = $state(false)
  let deletingPrompt = $state(false)

  const promptTextareaMaxHeight = 500

  let renderedLlmInput = $derived(
    renderFinalPrompt({
      template: workingTemplate,
      displayTitle: title,
      existingContent: editorContent,
      variableValues: { '기존 문서 내용': editorContent },
    }),
  )
  let hasPromptContent = $derived(workingTemplate.trim().length > 0)
  let canSavePrompt = $derived((getRLCONF()?.wgUserId ?? 0) > 0 && prompt.userId === getRLCONF()?.wgUserId)
  let isSysop = $derived((getRLCONF()?.wgUserGroups || []).includes('sysop'))
  let canDeletePrompt = $derived(prompt.id && (getRLCONF()?.wgUserId ?? 0) > 0 && (prompt.userId === getRLCONF()?.wgUserId || isSysop))
  let trimmedTitle = $derived(title.trim())
  let canSubmit = $derived(hasPromptContent && trimmedTitle.length > 0 && renderedLlmInput.trim().length > 0 && !submitting)
  let displayedTaskStatus = $derived(submitting ? 'Creating' : submittedTaskPhase === 'Completed' ? undefined : submittedTaskPhase)

  $effect(() => {
    const nextRunnerContextIdentity = `${prompt.id ?? ''}\u0000${prompt.title}\u0000${prompt.requestType}\u0000${title}\u0000${pageId ?? ''}`
    const nextPromptIdentity = `${prompt.id ?? ''}\u0000${prompt.title}\u0000${prompt.requestType}\u0000${prompt.content}\u0000${title}\u0000${pageId ?? ''}`
    if (nextPromptIdentity === lastRunnerIdentity) return

    const contextChanged = nextRunnerContextIdentity !== lastRunnerContextIdentity
    lastRunnerIdentity = nextPromptIdentity
    lastRunnerContextIdentity = nextRunnerContextIdentity
    if (!contextChanged && prompt.content === lastSavedContent) {
      lastSavedContent = undefined
      return
    }
    lastSavedContent = undefined
    if (!contextChanged && prompt.content === untrack(() => workingTemplate)) return

    workingTemplate = prompt.content
    submitError = ''
    resetResultState()
  })

  function handlePromptUpdate(value: string) {
    workingTemplate = value
    submitError = ''
  }

  async function handleSubmitClick() {
    if (!canSubmit) return

    submitting = true
    submitError = ''
    resetResultState()

    try {
      const data = await submit({
        pageId,
        title: trimmedTitle,
        promptTitle: prompt.title,
        requestType: prompt.requestType,
        llmInput: renderedLlmInput,
      })
      submittedTaskId = data.id
    } catch (error) {
      submitError = error instanceof Error && error.message ? error.message : 'AI 편집 등록에 실패했습니다.'
    } finally {
      submitting = false
    }
  }

  function resetResultState() {
    polling = false
    submittedTaskPhase = undefined
    submittedTaskId = undefined
    taskResetToken += 1
  }

  function closePreviewModal() {
    previewModalOpen = false
  }

  async function savePrompt() {
    if (!prompt.id) {
      showToast('저장할 프롬프트 정보가 없습니다.')
      return
    }
    if (!canSavePrompt) {
      showToast('자신의 프롬프트만 저장할 수 있습니다.')
      return
    }

    const savedContent = workingTemplate
    savingPrompt = true
    try {
      const [, err] = await httpy.post('/api/ai-prompts', {
        id: prompt.id,
        title: prompt.title,
        request_type: prompt.requestType,
        content: savedContent,
      })
      if (err) {
        showToast(err.message || '프롬프트 저장에 실패했습니다.')
        return
      }
      showToast('프롬프트를 저장했습니다.')
      lastSavedContent = savedContent
      void onPromptListChanged?.()
    } catch (error) {
      showToast(error instanceof Error && error.message ? error.message : '프롬프트 저장에 실패했습니다.')
    } finally {
      savingPrompt = false
    }
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
        content: workingTemplate,
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

  async function deletePrompt() {
    if (!prompt.id) {
      showToast('삭제할 프롬프트 정보가 없습니다.')
      return
    }
    if (!canDeletePrompt) {
      showToast('자신의 프롬프트만 삭제할 수 있습니다.')
      return
    }

    const ok = await showConfirm(`'${prompt.title}' 프롬프트를 삭제하시겠습니까?`)
    if (!ok) return

    deletingPrompt = true
    try {
      const [, err] = await httpy.delete(`/api/ai-prompts/${prompt.id}`)
      if (err) {
        showToast(err.message || '프롬프트 삭제에 실패했습니다.')
        return
      }
      showToast('프롬프트를 삭제했습니다.')
      void onPromptListChanged?.()
    } catch (error) {
      showToast(error instanceof Error && error.message ? error.message : '프롬프트 삭제에 실패했습니다.')
    } finally {
      deletingPrompt = false
    }
  }
</script>

<div class="flex h-full min-h-0 flex-col gap-3">
  {#if !hasPromptContent}
    <div class="rounded border border-a-amber-200 bg-a-amber-50 px-3 py-2 text-sm text-a-amber-700">프롬프트 내용이 없습니다.</div>
  {/if}

  <div class="flex min-h-0 flex-1 flex-col gap-3">
    <div class="flex min-h-0 flex-col gap-1">
      <div id="ai-edit-runner-prompt" class="relative min-h-40 font-mono text-xs leading-relaxed">
        <ZTextarea
          id="ai-edit-runner-prompt-textarea"
          class="bg-a-gray-50"
          modelValue={workingTemplate}
          placeholder="프롬프트 내용을 입력하세요."
          maxHeight={promptTextareaMaxHeight}
          onUpdateModelValue={handlePromptUpdate}
        />
        <div class="absolute right-4 top-1 flex items-center">
          {#if displayedTaskStatus}
            <div class="flex items-center text-xs text-a-slate-500">
              {#if submitting || polling}
                <ZSpinner size="0.875rem" extraClass="mr-1" />
              {/if}
              {displayedTaskStatus}
            </div>
          {/if}
          <CButton type="button" variant="ghost" size="small" disabled={!canSubmit || polling} onclick={() => void handleSubmitClick()}>
            <ZIcon path={mdiPlay} />
            Run
          </CButton>
          <CMenu>
            {#snippet trigger({ toggle })}
              <CButton type="button" variant="ghost" size="icon-sm" aria-label="프롬프트 메뉴" onclick={toggle}>
                <ZIcon path={mdiDotsVertical} />
              </CButton>
            {/snippet}
            {#snippet menu({ close })}
              <div class="flex flex-col items-stretch">
                <CButton
                  type="button"
                  variant="ghost"
                  class="w-full justify-start! rounded-none!"
                  disabled={!hasPromptContent}
                  onclick={() => {
                    previewModalOpen = true
                    close()
                  }}
                >
                  Render
                </CButton>
                <CButton
                  type="button"
                  variant="ghost"
                  class="w-full justify-start! rounded-none!"
                  disabled={!hasPromptContent || savingPrompt || !prompt.id || !canSavePrompt}
                  onclick={() => {
                    void savePrompt()
                    close()
                  }}
                >
                  {savingPrompt ? '저장 중...' : 'Save'}
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
                {#if canDeletePrompt}
                  <CButton
                    type="button"
                    variant="destructive"
                    class="w-full justify-start! rounded-none!"
                    disabled={deletingPrompt}
                    onclick={() => {
                      void deletePrompt()
                      close()
                    }}
                  >
                    {deletingPrompt ? '삭제 중...' : 'Delete'}
                  </CButton>
                {/if}
              </div>
            {/snippet}
          </CMenu>
        </div>
      </div>
    </div>
  </div>

  <div class="flex flex-col items-start gap-2">
    {#if submitError}
      <div class="text-sm text-a-red-500">{submitError}</div>
    {/if}
  </div>

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
</div>

<ZModal
  show={previewModalOpen}
  title="Rendered Prompt"
  okText="닫기"
  okVariant="default"
  showCancel={false}
  panelClass="max-w-[calc(100vw-2rem)] md:max-w-4xl"
  sectionClass="min-h-0 flex-1 overflow-hidden p-5"
  onOk={closePreviewModal}
  onCancel={closePreviewModal}
>
  <pre
    class="max-h-[calc(100dvh-14rem)] overflow-auto whitespace-pre-wrap rounded border border-a-slate-200 bg-a-slate-50 p-3 font-mono text-xs text-a-slate-800">{renderedLlmInput}</pre>
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
