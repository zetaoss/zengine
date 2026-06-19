<svelte:options runes={true} />

<script lang="ts">
  import type {
    AIEditExistingContentResult,
    AIEditPromptForRunner,
    AIEditRunnerLayout,
    AIEditRunnerSubmitPayload,
    ExistingContentState,
  } from '$shared/types/aiEdit'
  import CButton from '$shared/ui/CButton.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'

  import AiEditPromptPreview from './AiEditPromptPreview.svelte'
  import { extractTemplateVariables, getPromptParts, renderFinalPrompt, usesTemplateVariable } from './prompt/promptRenderer'

  type SubmitHandler = (payload: AIEditRunnerSubmitPayload) => Promise<void>
  type ExistingContentLoader = (title: string) => Promise<AIEditExistingContentResult>

  let {
    prompt,
    initialTitle = '',
    initialPageId = undefined,
    loadExistingContent = undefined,
    submit,
    layout = 'full',
    disabled = false,
    disabledReason = '',
    titleReadonly = false,
    hideTitle = false,
    externalContent = undefined,
    externalContentDirty = false,
    hideReservedStatus = false,
    class: className = '',
    externalContentVersion = 0,
    onContentResetRequest = undefined,
  }: {
    prompt: AIEditPromptForRunner
    initialTitle?: string
    initialPageId?: number
    loadExistingContent?: ExistingContentLoader
    submit: SubmitHandler
    layout?: AIEditRunnerLayout
    disabled?: boolean
    disabledReason?: string
    titleReadonly?: boolean
    hideTitle?: boolean
    externalContent?: string
    externalContentDirty?: boolean
    hideReservedStatus?: boolean
    class?: string
    externalContentVersion?: number
    onContentResetRequest?: () => void
  } = $props()

  let title = $state('')
  let pageId = $state<number | undefined>(undefined)
  let displayTitle = $state('')
  let displayTitleTouched = $state(false)
  let variableValues = $state<Record<string, string>>({})
  let existingContentState = $state<ExistingContentState>('idle')
  let existingContent = $state('')
  let verifiedTitle = $state('')
  let existingContentError = $state('')
  let loadingTitle = $state('')
  let submitting = $state(false)
  let submitError = $state('')
  let workingTemplate = $state('')
  let lastRunnerIdentity = $state('')
  let lastVariableKey = $state('')
  let promptPreviewRefreshKey = $state(0)
  let lastExternalContentResetVersion = 0
  let lastExistingContentLoadStartedAt = 0
  let loadingStartedAt = $state<number | undefined>(undefined)
  let loadSerial = 0

  const titleCheckDebounceMs = 650
  const titleCheckMinIntervalMs = 1500
  const existingContentLoadTimeoutMs = 8000
  const legacyAutoVariableNames = new Set(['기존문서', '문서내용', '표시제목'])

  let templateVariables = $derived(extractTemplateVariables(workingTemplate))
  let variableKey = $derived(templateVariables.map((variable) => variable.name).join('\u0000'))
  let needsExistingContent = $derived(prompt.requestType === 'edit' && usesTemplateVariable(workingTemplate, '기존 문서 내용'))
  let requiresTitleCheck = $derived((prompt.requestType === 'create' || prompt.requestType === 'edit') && externalContent === undefined)
  let renderedLlmInput = $derived(
    renderFinalPrompt({
      template: workingTemplate,
      displayTitle,
      existingContent: needsExistingContent ? existingContent : '',
      variableValues,
    }),
  )
  let previewParts = $derived(
    getPromptParts({
      template: workingTemplate,
      displayTitle,
      existingContent: needsExistingContent ? existingContent : '',
      variableValues,
      rawMode: false,
    }),
  )
  let reservedDefaults = $derived({
    제목: verifiedTitle,
    '기존 문서 내용': existingContent,
  })

  let legacyVariables = $derived(templateVariables.filter((variable) => legacyAutoVariableNames.has(variable.name)))
  let hasPromptContent = $derived(workingTemplate.trim().length > 0)
  let trimmedTitle = $derived(title.trim())
  let titleRequirementSatisfied = $derived(
    !requiresTitleCheck ||
      (prompt.requestType === 'edit' && existingContentState === 'ready') ||
      (prompt.requestType === 'create' && existingContentState === 'available') ||
      externalContent !== undefined,
  )
  let canSubmit = $derived(
    hasPromptContent &&
      trimmedTitle.length > 0 &&
      renderedLlmInput.trim().length > 0 &&
      !disabled &&
      !submitting &&
      titleRequirementSatisfied,
  )
  let rootClass = $derived(
    `flex min-h-0 flex-col gap-3 border border-border p-4 ${layout === 'compact' ? 'h-full' : 'rounded-lg bg-background'} ${className}`,
  )
  let bodyClass = $derived('flex min-h-0 flex-1 flex-col gap-3')
  let titleLabel = $derived(
    prompt.requestType === 'create'
      ? '생성할 문서 제목'
      : prompt.requestType === 'edit'
        ? '편집할 문서 제목'
        : '문서 제목',
  )
  let titlePlaceholder = $derived(`${titleLabel}을 입력하세요.`)

  $effect(() => {
    const nextPromptIdentity = `${prompt.id ?? ''}\u0000${prompt.title}\u0000${prompt.requestType}\u0000${prompt.content}\u0000${initialTitle}\u0000${initialPageId ?? ''}`
    if (nextPromptIdentity === lastRunnerIdentity) return

    const nextVariables = extractTemplateVariables(prompt.content)
    const nextVariableKey = nextVariables.map((variable) => variable.name).join('\u0000')
    const nextRequiresTitleCheck = (prompt.requestType === 'create' || prompt.requestType === 'edit') && externalContent === undefined

    lastRunnerIdentity = nextPromptIdentity
    lastVariableKey = nextVariableKey
    workingTemplate = prompt.content
    title = initialTitle
    pageId = initialPageId
    displayTitle = initialTitle
    displayTitleTouched = false
    const nextVariableValues = buildVariableValues(nextVariables)
    if (usesTemplateVariable(prompt.content, '기존 문서 내용')) {
      nextVariableValues['기존 문서 내용'] = externalContent ?? ''
    }
    variableValues = nextVariableValues
    existingContent = externalContent ?? ''
    verifiedTitle = externalContent !== undefined ? initialTitle : ''
    existingContentError = ''
    loadingTitle = ''
    loadingStartedAt = undefined
    existingContentState =
      externalContent !== undefined ? 'ready' : nextRequiresTitleCheck ? (initialTitle.trim() ? 'stale' : 'idle') : 'not_required'
    submitError = ''
  })

  $effect(() => {
    // This effect runs when the parent component signals a content reset by bumping the version.
    if (externalContentVersion === 0 || externalContentVersion === lastExternalContentResetVersion) return
    lastExternalContentResetVersion = externalContentVersion
    if (externalContent !== undefined) {
      existingContent = externalContent
      verifiedTitle = trimmedTitle
      existingContentState = 'ready'
      existingContentError = ''
      loadingTitle = ''
      loadingStartedAt = undefined
      variableValues['기존 문서 내용'] = externalContent
      promptPreviewRefreshKey += 1
    }
  })

  $effect(() => {
    if (variableKey === lastVariableKey) return
    lastVariableKey = variableKey
    const nextValues: Record<string, string> = {}
    for (const variable of templateVariables) {
      nextValues[variable.name] = variableValues[variable.name] ?? ''
    }
    if (usesTemplateVariable(workingTemplate, '기존 문서 내용')) {
      nextValues['기존 문서 내용'] = variableValues['기존 문서 내용'] ?? existingContent
    }
    variableValues = nextValues
  })

  $effect(() => {
    if (externalContent !== undefined) {
      existingContentState = 'ready'
      verifiedTitle = trimmedTitle
      existingContentError = ''
      loadingTitle = ''
      loadingStartedAt = undefined
      return
    }

    if (!requiresTitleCheck) {
      existingContentState = 'not_required'
      existingContent = ''
      verifiedTitle = ''
      existingContentError = ''
      loadingTitle = ''
      loadingStartedAt = undefined
      return
    }

    const currentTitle = trimmedTitle
    if (!currentTitle) {
      existingContentState = 'idle'
      existingContent = ''
      verifiedTitle = ''
      existingContentError = ''
      loadingTitle = ''
      loadingStartedAt = undefined
      return
    }

    if (verifiedTitle === currentTitle && existingContentState !== 'stale' && existingContentState !== 'loading') return
    if (existingContentState === 'loading' && loadingTitle === currentTitle) return

    if (verifiedTitle && verifiedTitle !== currentTitle && existingContentState !== 'stale') {
      existingContentState = 'stale'
      existingContent = ''
      existingContentError = ''
    }

    const timer = setTimeout(() => {
      void loadExistingContentFor(currentTitle)
    }, getTitleCheckDelay())

    return () => clearTimeout(timer)
  })

  // NOTE: This effect was causing an infinite loop. It's a nice-to-have UX feature
  // to sync the display title with a server-normalized title, but it creates a
  // reactive loop that freezes the browser. Disabling it to fix the freeze.
  // $effect(() => {
  //   if (prompt.requestType !== 'create' || existingContentState !== 'available' || !verifiedTitle) return
  //   syncDisplayTitleFromTarget(verifiedTitle)
  // })

  $effect(() => {
    if (existingContentState !== 'loading' || !loadingTitle || loadingStartedAt === undefined) return

    const serial = loadSerial
    const titleAtStart = loadingTitle
    const remainingMs = Math.max(0, existingContentLoadTimeoutMs - (Date.now() - loadingStartedAt))
    const timer = setTimeout(() => {
      if (existingContentState !== 'loading' || loadingTitle !== titleAtStart || loadSerial !== serial) return

      loadSerial += 1
      existingContent = ''
      verifiedTitle = titleAtStart
      pageId = undefined
      existingContentError = '문서 확인 시간이 초과되었습니다. 잠시 후 다시 시도하세요.'
      existingContentState = 'error'
      loadingTitle = ''
      loadingStartedAt = undefined
    }, remainingMs)

    return () => clearTimeout(timer)
  })

  function buildVariableValues(variables = templateVariables) {
    const nextValues: Record<string, string> = {}
    for (const variable of variables) {
      nextValues[variable.name] = variableValues[variable.name] ?? ''
    }
    return nextValues
  }

  function syncDisplayTitleFromTarget(pageTitle: string) {
    displayTitle = pageTitle
    promptPreviewRefreshKey += 1
  }

  function handleTitleInput(event: Event) {
    const nextTitle = (event.currentTarget as HTMLInputElement).value
    title = nextTitle
    pageId = undefined
    submitError = ''

    if (!displayTitleTouched) {
      syncDisplayTitleFromTarget(nextTitle)
    }

    if (requiresTitleCheck) {
      if (existingContentState === 'loading') loadSerial += 1
      existingContent = ''
      verifiedTitle = ''
      existingContentError = ''
      loadingTitle = ''
      loadingStartedAt = undefined
      existingContentState = nextTitle.trim() ? 'stale' : 'idle'
    }
  }

  function handleDisplayTitleInput(event: Event) {
    displayTitle = (event.currentTarget as HTMLInputElement).value
    displayTitleTouched = true
    submitError = ''
  }

  function handleInteractiveInput() {
    submitError = ''
  }

  function handleTitleReset() {
    if (!verifiedTitle) return
    displayTitle = verifiedTitle
    displayTitleTouched = false
    promptPreviewRefreshKey += 1
  }

  async function loadExistingContentFor(pageTitle: string) {
    if (!loadExistingContent) {
      existingContentState = 'error'
      existingContentError = '기존 문서를 불러올 수 없습니다.'
      return
    }

    const serial = ++loadSerial
    const startedAt = Date.now()
    lastExistingContentLoadStartedAt = startedAt
    loadingStartedAt = startedAt
    loadingTitle = pageTitle
    existingContentState = 'loading'
    existingContentError = ''

    try {
      const result = await withTimeout(loadExistingContent(pageTitle), existingContentLoadTimeoutMs)
      if (serial !== loadSerial || trimmedTitle !== pageTitle) return

      if (prompt.requestType === 'create') {
        const resolvedTitle = result.title || pageTitle
        existingContent = ''
        verifiedTitle = resolvedTitle
        pageId = undefined
        existingContentState = 'error'
        existingContentError = '이미 존재하는 문서입니다.'
        return
      }

      const resolvedTitle = result.title || pageTitle
      existingContent = result.content
      verifiedTitle = resolvedTitle
      if (!displayTitleTouched) {
        syncDisplayTitleFromTarget(resolvedTitle)
      }
      pageId = result.pageId ?? initialPageId
      existingContentState = 'ready'

      // Keep the editable reserved variable value in sync with the fetched source.
      variableValues['기존 문서 내용'] = result.content
    } catch (error) {
      if (serial !== loadSerial || trimmedTitle !== pageTitle) return

      const resolvedErrorTitle = getErrorTitle(error) ?? pageTitle
      existingContent = ''
      verifiedTitle = resolvedErrorTitle
      pageId = undefined
      existingContentError = getErrorMessage(error)

      if (error instanceof Error && error.name === 'InvalidTitleError') {
        existingContentState = 'invalid'
      } else if (error instanceof Error && error.name === 'NotFoundError') {
        if (prompt.requestType === 'create') {
          existingContentState = 'available'
          syncDisplayTitleFromTarget(resolvedErrorTitle)
        } else {
          existingContentState = 'not_found'
        }
      } else {
        existingContentState = 'error'
      }
    } finally {
      if (serial === loadSerial) {
        loadingTitle = ''
        loadingStartedAt = undefined
      }
    }
  }

  function getTitleCheckDelay() {
    const elapsedSinceLastLoad = Date.now() - lastExistingContentLoadStartedAt
    const throttleDelay = Math.max(0, titleCheckMinIntervalMs - elapsedSinceLastLoad)
    return Math.max(titleCheckDebounceMs, throttleDelay)
  }

  function withTimeout<T>(promise: Promise<T>, timeoutMs: number): Promise<T> {
    let timeoutId: ReturnType<typeof setTimeout> | undefined
    const timeout = new Promise<never>((_, reject) => {
      timeoutId = setTimeout(() => {
        const error = new Error('문서 확인 시간이 초과되었습니다. 잠시 후 다시 시도하세요.')
        error.name = 'TimeoutError'
        reject(error)
      }, timeoutMs)
    })

    return Promise.race([promise, timeout]).finally(() => {
      if (timeoutId !== undefined) clearTimeout(timeoutId)
    })
  }

  function getErrorMessage(error: unknown) {
    if (error instanceof Error && error.message) return error.message
    return '기존 문서를 불러오지 못했습니다.'
  }

  function getErrorTitle(error: unknown) {
    if (!(error instanceof Error) || !('title' in error)) return undefined
    const title = (error as Error & { title?: unknown }).title
    return typeof title === 'string' && title.trim() ? title : undefined
  }

  let statusText = $derived.by(() => {
    if (!requiresTitleCheck || hideReservedStatus) return ''
    if (existingContentState === 'idle') return ''
    if (existingContentState === 'loading' || existingContentState === 'stale') return '로딩 중'

    if (prompt.requestType === 'create') {
      if (existingContentState === 'available') return '생성 가능'
      if (existingContentState === 'error' && existingContentError.includes('이미 존재')) return '이미 있음'
      if (existingContentState === 'ready') return '이미 있음'
    } else {
      if (existingContentState === 'ready') return '로딩 완료'
      if (existingContentState === 'not_found') return '문서 없음'
    }

    if (existingContentState === 'invalid') return '유효하지 않음'
    if (existingContentState === 'error') return '오류'
    return ''
  })

  let statusBadgeClass = $derived.by(() => {
    if (statusText === '로딩 완료' || statusText === '생성 가능') {
      return 'text-a-emerald-600'
    }
    if (statusText === '로딩 중') {
      return 'text-a-blue-600'
    }
    if (statusText === '이미 있음' || statusText === '문서 없음' || statusText === '유효하지 않음' || statusText === '오류') {
      return 'text-a-red-600'
    }
    return 'text-a-slate-600'
  })

  let inputBorderClass = $derived.by(() => {
    if (statusText === '로딩 완료' || statusText === '생성 가능') {
      return 'border-a-emerald-500 focus:border-a-emerald-600'
    }
    if (statusText === '로딩 중') {
      return 'border-a-blue-500 focus:border-a-blue-600'
    }
    if (statusText === '이미 있음' || statusText === '문서 없음' || statusText === '유효하지 않음' || statusText === '오류') {
      return 'border-a-red-500 focus:border-a-red-600'
    }
    return 'border-a-slate-300 focus:border-a-blue-500'
  })

  async function handleSubmitClick() {
    if (!canSubmit) return

    submitting = true
    submitError = ''

    try {
      await submit({
        pageId,
        title: trimmedTitle,
        promptTitle: prompt.title,
        requestType: prompt.requestType,
        llmInput: renderedLlmInput,
      })
    } catch (error) {
      submitError = error instanceof Error && error.message ? error.message : 'AI 편집 등록에 실패했습니다.'
    } finally {
      submitting = false
    }
  }
</script>

<div class={rootClass}>
  {#if disabled && disabledReason}
    <div class="rounded border border-a-slate-200 bg-a-slate-50 px-3 py-2 text-sm text-a-slate-600">{disabledReason}</div>
  {/if}

  {#if !hasPromptContent}
    <div class="rounded border border-a-amber-200 bg-a-amber-50 px-3 py-2 text-sm text-a-amber-700">프롬프트 내용이 없습니다.</div>
  {/if}

  <div class={bodyClass}>
    <div class="flex min-w-0 flex-col gap-3">
      {#if !hideTitle}
        <div class="grid gap-2 {layout === 'full' ? 'grid-cols-1' : 'md:grid-cols-2'}">
          <div class="flex min-w-0 flex-col gap-1 text-sm text-a-slate-700">
            <span class="font-semibold">{titleLabel}</span>
            <div class="flex items-center gap-2">
              <input
                type="text"
                class={`min-w-0 flex-1 rounded border px-2 py-1 text-sm outline-none transition-colors ${inputBorderClass}`}
                value={title}
                readonly={titleReadonly}
                oninput={handleTitleInput}
                placeholder={titlePlaceholder}
              />
              {#if statusText}
                <div class={`flex shrink-0 items-center gap-1 text-xs font-semibold ${statusBadgeClass}`}>
                  {#if existingContentState === 'loading'}
                    <ZSpinner size="0.875rem" extraClass="mr-0" />
                  {/if}
                  <span>{statusText}</span>
                </div>
              {/if}
            </div>
          </div>

          {#if layout === 'compact'}
            <label class="flex min-w-0 flex-col gap-1 text-sm text-a-slate-700">
              <span class="font-semibold">표시용 제목 ({`{제목}`})</span>
              <input
                type="text"
                class="min-w-0 rounded border border-a-slate-300 px-2 py-1 text-sm"
                value={displayTitle}
                oninput={handleDisplayTitleInput}
                placeholder="프롬프트에 들어갈 제목"
              />
            </label>
          {/if}
        </div>
      {/if}

      {#if legacyVariables.length > 0}
        <div class="rounded border border-a-amber-200 bg-a-amber-50 px-3 py-2 text-sm text-a-amber-700">
          자동 변수로 처리되지 않는 이름이 있습니다: {legacyVariables.map((variable) => `{${variable.name}}`).join(', ')}
        </div>
      {/if}
    </div>

    <div class="flex min-h-0 flex-col gap-1">
      <label for="ai-edit-runner-prompt-preview" class="text-sm font-semibold text-a-slate-700">프롬프트</label>
      <AiEditPromptPreview
        id="ai-edit-runner-prompt-preview"
        bind:template={workingTemplate}
        parts={previewParts}
        bind:displayTitle
        bind:variableValues
        {reservedDefaults}
        isExternalContent={externalContent !== undefined}
        {externalContentDirty}
        refreshKey={promptPreviewRefreshKey}
        editable={true}
        oninput={handleInteractiveInput}
        onTitleReset={handleTitleReset}
        onContentResetRequest={onContentResetRequest}
        class={layout === 'compact' ? '' : 'min-h-[500px]'}
      />
    </div>
  </div>

  <div class="flex flex-col items-center gap-2">
    <CButton variant="outline" disabled={!canSubmit} onclick={() => void handleSubmitClick()}>
      {submitting ? '등록 중' : 'AI 편집 등록'}
    </CButton>
    {#if submitError}
      <div class="text-sm text-a-red-500">{submitError}</div>
    {/if}
  </div>
</div>
