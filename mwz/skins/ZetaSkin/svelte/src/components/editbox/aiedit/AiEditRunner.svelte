<svelte:options runes={true} />

<script lang="ts">
  import { mdiClose, mdiContentCopy, mdiPlay, mdiPlus } from '@mdi/js'
  import type { Snippet } from 'svelte'

  import CButton from '$shared/ui/CButton.svelte'
  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZModal from '$shared/ui/ZModal.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import ZTextarea from '$shared/ui/ZTextarea.svelte'
  import httpy from '$shared/utils/httpy'

  import AiEditTaskResults from './AiEditTaskResults.svelte'
  import type {
    AIEditMessageTemplates,
    AIEditPromptForRunner,
    AIEditRenderedMessage,
    AIEditRequestPayload,
    AIEditRequestType,
    AIEditStoreResult,
    AIEditTaskPhase,
    AIEditUserInput,
    AIEditUserInputValue,
  } from './aiEditTypes'

  interface TextSegment {
    text: string
    kind: 'plain' | 'variable' | 'control' | 'value'
  }

  let {
    prompt,
    message,
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
    message: AIEditMessageTemplates
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
  let previewOpen = $state(false)
  let previewLoading = $state(false)
  let previewError = $state('')
  let previewMessages = $state<AIEditRenderedMessage[]>([])
  let previewPayload = $state<AIEditRequestPayload | null>(null)
  let saveAsModalOpen = $state(false)
  let saveAsTitle = $state('')
  let savingPromptAs = $state(false)

  let hasPromptContent = $derived(prompt.spec.trim().length > 0 && !prompt.error)
  let trimmedTitle = $derived(title.trim())
  let requiredFilled = $derived(userInputs.every((input) => !input.is_required || filled(values[input.name])))
  let canPreview = $derived(hasPromptContent && trimmedTitle.length > 0 && !submitting && !polling && !previewLoading)
  let fetchingBacklinks = $state(false)
  let backlinksFetchStartedFor = $state('')
  let canSubmit = $derived(hasPromptContent && trimmedTitle.length > 0 && requiredFilled && !submitting)

  // An input starts at its default; a value kept from another template that fits it stays.
  $effect(() => {
    const missing = userInputs.filter((input) => input.default !== undefined && !fits(input, values[input.name]))
    if (missing.length)
      values = { ...values, ...Object.fromEntries(missing.map((input) => [input.name, input.default as AIEditUserInputValue])) }
  })
  let displayedTaskStatus = $derived(submitting ? 'Creating' : submittedTaskPhase === 'Completed' ? undefined : submittedTaskPhase)

  $effect(() => {
    const pageTitle = trimmedTitle
    if (!pageTitle || !userInputs.some((input) => input.name === 'backlinks') || backlinksFetchStartedFor === pageTitle) return
    if (textsValue('backlinks').some((value) => value.trim())) {
      backlinksFetchStartedFor = pageTitle
      return
    }
    backlinksFetchStartedFor = pageTitle
    void fetchBacklinks('backlinks', pageTitle)
  })

  $effect(() => {
    if (!previewOpen) return
    const request = makePreviewRequest(payload())
    let active = true
    const timer = setTimeout(async () => {
      previewLoading = true
      previewError = ''
      previewPayload = request
      try {
        const messages = await preview(request)
        if (active) previewMessages = messages
      } catch (error) {
        if (active) previewError = error instanceof Error && error.message ? error.message : '미리보기를 만들지 못했습니다.'
      } finally {
        if (active) previewLoading = false
      }
    }, 250)
    return () => {
      active = false
      clearTimeout(timer)
    }
  })

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

  // backlinks is filled by asking the wiki which pages link to the title; the list can then be edited. The answer
  // is dropped when the title or the list changed while it was on its way, so it never replaces what was typed.
  async function fetchBacklinks(name: string, pageTitle: string) {
    if (!pageTitle) return
    fetchingBacklinks = true
    const before = JSON.stringify(textsValue(name))
    try {
      const [data, err] = await httpy.get<{ titles: string[] }>('/api/ai-edit/backlinks', { title: pageTitle })
      if (err) {
        showToast(err.message || '역링크를 가져오지 못했습니다.')
        return
      }
      if (pageTitle !== trimmedTitle || JSON.stringify(textsValue(name)) !== before) return
      setTexts(name, data?.titles ?? [])
    } finally {
      fetchingBacklinks = false
    }
  }

  // A user input's name as the run form shows it: someName as "Some name".
  function inputLabel(name: string) {
    const words = name.replace(/([a-z0-9])([A-Z])/g, '$1 $2').toLowerCase()
    return words.charAt(0).toUpperCase() + words.slice(1)
  }

  function messageForRole(messages: Array<{ role: string; content: string }>, role: string) {
    return messages
      .filter((m) => m.role === role)
      .map((m) => m.content)
      .join('\n\n')
  }

  let promptMessages = $derived(
    (['system', 'user'] as const).filter((role) => message[role]?.trim()).map((role) => ({ role, content: message[role] ?? '' })),
  )
  let messageRows = $derived(promptMessages.map(({ role }) => ({ role })))

  // A value fits an input of its type: a value kept from another template with the same name but another type (or,
  // for a choice, not one of its options) is not sent.
  function fits(input: AIEditUserInput, v: AIEditUserInputValue | undefined) {
    if (input.type === 'bool') return typeof v === 'boolean'
    if (input.type === 'textList') return Array.isArray(v)
    if (input.type === 'choice') return typeof v === 'string' && (input.json_schema?.enum ?? []).includes(v)
    return typeof v === 'string'
  }

  // A required input is filled with a non-blank text or a list with a non-blank item.
  function filled(v: AIEditUserInputValue | undefined) {
    if (typeof v === 'string') return v.trim() !== ''
    if (Array.isArray(v)) return v.some((x) => x.trim() !== '')
    return v !== undefined
  }

  function payload(): AIEditRequestPayload {
    const userValues: Record<string, AIEditUserInputValue> = {}
    for (const input of userInputs) {
      const v = values[input.name]
      if (v !== undefined && fits(input, v)) userValues[input.name] = v
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

  function openPreview() {
    previewOpen = true
    previewMessages = []
    previewError = ''
  }

  function closePreview() {
    previewOpen = false
    previewLoading = false
  }

  function makePreviewRequest(request: AIEditRequestPayload): AIEditRequestPayload {
    return {
      ...request,
      title: shortPreviewValue(request.title),
      wikitext: shortPreviewValue(request.wikitext),
      userInputs: Object.fromEntries(
        Object.entries(request.userInputs).map(([name, value]) => [
          name,
          typeof value === 'string' ? shortPreviewValue(value) : Array.isArray(value) ? value.map(shortPreviewValue) : value,
        ]),
      ),
    }
  }

  function shortPreviewValue(value: string) {
    return value.length > 100 ? `${value.slice(0, 100)}...` : value
  }

  function originalSegments(content: string): TextSegment[] {
    const segments: TextSegment[] = []
    const expression = /\{\{[^{}]*\}\}|\{%[\s\S]*?%\}/g
    let cursor = 0
    for (const match of content.matchAll(expression)) {
      const token = match[0]
      let kind: TextSegment['kind'] = 'plain'
      if (token.startsWith('{%')) {
        kind = 'control'
      } else {
        const name = token
          .slice(2, -2)
          .trim()
          .match(/^([A-Za-z][A-Za-z0-9_.]*)/)?.[1]
        if (name && name !== name.toUpperCase() && !token.includes('|') && !token.includes(':')) kind = 'variable'
      }
      if (kind === 'plain' || match.index === undefined) continue
      if (match.index > cursor) segments.push({ text: content.slice(cursor, match.index), kind: 'plain' })
      segments.push({ text: token, kind })
      cursor = match.index + token.length
    }
    if (cursor < content.length) segments.push({ text: content.slice(cursor), kind: 'plain' })
    return segments.length ? segments : [{ text: content, kind: 'plain' }]
  }

  function renderedSegments(content: string): TextSegment[] {
    if (!previewPayload) return [{ text: content, kind: 'plain' }]
    const values = [
      previewPayload.title,
      ...(previewPayload.requestType === 'update' ? [previewPayload.wikitext] : []),
      ...Object.values(previewPayload.userInputs).flatMap((value) =>
        typeof value === 'string' ? [value] : Array.isArray(value) ? value : [],
      ),
    ]
      .filter(Boolean)
      .filter((value, index, all) => all.indexOf(value) === index)
      .sort((a, b) => b.length - a.length)

    const segments: TextSegment[] = []
    let cursor = 0
    while (cursor < content.length) {
      let nextIndex = -1
      let nextValue = ''
      for (const value of values) {
        const index = content.indexOf(value, cursor)
        if (index >= 0 && (nextIndex < 0 || index < nextIndex || (index === nextIndex && value.length > nextValue.length))) {
          nextIndex = index
          nextValue = value
        }
      }
      if (nextIndex < 0) {
        segments.push({ text: content.slice(cursor), kind: 'plain' })
        break
      }
      if (nextIndex > cursor) segments.push({ text: content.slice(cursor, nextIndex), kind: 'plain' })
      segments.push({ text: nextValue, kind: 'value' })
      cursor = nextIndex + nextValue.length
    }
    return segments.length ? segments : [{ text: content, kind: 'plain' }]
  }

  function resetResultState() {
    polling = false
    submittedTaskPhase = undefined
    submittedTaskId = undefined
    taskResetToken += 1
  }

  function openSaveAsModal() {
    saveAsTitle = `${prompt.title} (복제본)`
    saveAsModalOpen = true
  }

  // The spec with its name replaced; a JSON string is a YAML double-quoted scalar.
  function renamedSpec(spec: string, name: string) {
    return spec.replace(/^name:.*$/m, `name: ${JSON.stringify(name)}`)
  }

  async function savePromptAs() {
    const newTitle = saveAsTitle.trim()
    if (!newTitle) return

    savingPromptAs = true
    try {
      const [, err] = await httpy.post('/api/ai-prompts', {
        request_type: prompt.requestType,
        spec: renamedSpec(prompt.spec, newTitle),
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
    <div class="flex flex-wrap items-center gap-2">
      <div class="min-w-0 flex-1">{@render promptHeader?.()}</div>
      <div class="flex shrink-0 items-center gap-1">
        <CButton
          type="button"
          variant="outline"
          size="small"
          disabled={!previewOpen && !canPreview}
          style={previewOpen ? 'color: var(--color-progressive)' : undefined}
          onclick={() => {
            if (previewOpen) closePreview()
            else void openPreview()
          }}
        >
          {#if previewLoading}<ZSpinner size="0.875rem" extraClass="mr-1" />{/if}
          미리보기
        </CButton>
        <CButton type="button" variant="ghost" size="small" disabled={!hasPromptContent || savingPromptAs} onclick={openSaveAsModal}>
          <ZIcon path={mdiContentCopy} />
          복제
        </CButton>
      </div>
    </div>
    {#if prompt.error}
      <div class="rounded border border-a-amber-200 bg-a-amber-50 px-3 py-2 text-sm text-a-amber-700">
        이 템플릿은 실행할 수 없습니다: {prompt.error}
      </div>
    {:else if !hasPromptContent}
      <div class="rounded border border-a-amber-200 bg-a-amber-50 px-3 py-2 text-sm text-a-amber-700">프롬프트 내용이 없습니다.</div>
    {/if}

    <div class={previewOpen ? 'grid min-h-0 flex-1 gap-3 md:grid-cols-2' : 'flex min-h-0 flex-1 flex-col gap-3'}>
      <div class="flex min-h-0 min-w-0 flex-col gap-3">
        <section class="grid h-[45rem] shrink-0 auto-rows-fr gap-2 overflow-hidden" aria-label="프롬프트 내용">
          {#if promptMessages.length}
            {#each messageRows as row (row.role)}
              <article class="relative flex min-h-0 min-w-0 flex-col overflow-hidden rounded border border-border bg-card">
                <span
                  class="pointer-events-none absolute -top-px right-[13px] rounded px-1.5 py-0.5 font-mono text-xs text-muted-foreground/70"
                  >{row.role}</span
                >
                <pre
                  class="min-h-0 min-w-0 flex-1 overflow-auto p-3 pr-16 whitespace-pre-wrap font-sans text-sm leading-relaxed text-foreground">{#each originalSegments(messageForRole(promptMessages, row.role)) as segment, index (`${row.role}-${index}`)}<span
                      class={segment.kind === 'control' ? 'rounded-sm bg-muted px-1 text-accent-foreground' : ''}
                      style={segment.kind === 'variable' ? 'color: var(--color-progressive)' : undefined}>{segment.text}</span
                    >{/each}</pre>
              </article>
            {/each}
          {:else}
            <div class="rounded border border-dashed border-border bg-card px-3 py-4 text-sm text-muted-foreground md:col-span-2">
              프롬프트 내용을 불러오는 중입니다…
            </div>
          {/if}
        </section>

        {#if userInputs.length > 0}
          <div class="flex min-h-0 flex-1 flex-col gap-3 overflow-auto text-sm">
            {#each userInputs as input (input.name)}
              {@const label = inputLabel(input.name) + (input.is_required ? ' *' : '')}
              {@const placeholder = input.ui?.placeholder}
              {@const long = input.ui?.variant === 'longText'}
              <div class="flex flex-col gap-1" title={input.description}>
                {#if input.type === 'bool'}
                  <label class="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={values[input.name] === true}
                      onchange={(e) => (values = { ...values, [input.name]: e.currentTarget.checked })}
                    />
                    {label}
                  </label>
                {:else if input.type === 'textList'}
                  <div class="flex flex-col gap-2">
                    <div class="text-xs">
                      <span class="font-semibold">{label}</span>
                    </div>
                    {#each textsValue(input.name) as item, i (i)}
                      <div class="flex items-start gap-1">
                        <div class="min-w-0 flex-1 text-xs">
                          <ZTextarea
                            id={`ai-edit-input-${input.name}-${i}`}
                            modelValue={item}
                            placeholder={placeholder ?? `${label} ${i + 1}`}
                            class={long ? 'min-h-32' : ''}
                            maxHeight={long ? 480 : 240}
                            onUpdateModelValue={(v: string) =>
                              setTexts(
                                input.name,
                                textsValue(input.name).map((x, j) => (j === i ? v : x)),
                              )}
                          />
                        </div>
                        <CButton
                          type="button"
                          variant="ghost"
                          size="icon-sm"
                          aria-label={`${label} ${i + 1} 빼기`}
                          onclick={() =>
                            setTexts(
                              input.name,
                              textsValue(input.name).filter((_, j) => j !== i),
                            )}
                        >
                          <ZIcon path={mdiClose} />
                        </CButton>
                      </div>
                    {/each}
                    <div class="flex items-center gap-1">
                      <CButton
                        type="button"
                        variant="outline"
                        size="small"
                        onclick={() => setTexts(input.name, [...textsValue(input.name), ''])}
                      >
                        <ZIcon path={mdiPlus} />
                        {inputLabel(input.name)} 추가
                      </CButton>
                      {#if input.name === 'backlinks'}
                        {#if fetchingBacklinks}<span class="flex items-center text-xs text-muted-foreground"
                            ><ZSpinner size="0.875rem" extraClass="mr-1" />역링크 불러오는 중…</span
                          >{/if}
                      {/if}
                    </div>
                  </div>
                {:else}
                  <div class="flex flex-col gap-1 text-xs">
                    <div>
                      <span class="font-semibold">{label}</span>
                    </div>
                    {#if input.type === 'choice'}
                      <select
                        id={`ai-edit-input-${input.name}`}
                        class="border-border h-8 rounded border bg-background px-2"
                        value={textValue(input.name)}
                        onchange={(e) => setText(input.name, e.currentTarget.value)}
                      >
                        {#if input.default === undefined}
                          <option value="">{placeholder ?? '선택'}</option>
                        {/if}
                        {#each input.json_schema?.enum ?? [] as option (option)}
                          <option value={option}>{option}</option>
                        {/each}
                      </select>
                    {:else if input.type === 'text'}
                      <ZTextarea
                        id={`ai-edit-input-${input.name}`}
                        modelValue={textValue(input.name)}
                        {placeholder}
                        class="min-h-32"
                        maxHeight={480}
                        onUpdateModelValue={(v: string) => setText(input.name, v)}
                      />
                    {:else}
                      <input
                        id={`ai-edit-input-${input.name}`}
                        class="border-border h-8 rounded border bg-background px-2"
                        {placeholder}
                        value={textValue(input.name)}
                        oninput={(e) => setText(input.name, e.currentTarget.value)}
                      />
                    {/if}
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        {/if}

        <div class="flex shrink-0 flex-col items-start gap-2">
          {#if submitError}
            <div class="text-sm text-a-red-500">{submitError}</div>
          {/if}
          {#if displayedTaskStatus}
            <div class="flex items-center text-xs text-a-slate-500">
              {#if submitting || polling}<ZSpinner size="0.875rem" extraClass="mr-1" />{/if}
              {displayedTaskStatus}
            </div>
          {/if}
          <CButton type="button" variant="subtle" size="small" disabled={!canSubmit || polling} onclick={() => void handleSubmitClick()}>
            <ZIcon path={mdiPlay} />
            실행
          </CButton>
        </div>
      </div>

      {#if previewOpen}
        <section class="grid h-[45rem] min-h-0 min-w-0 shrink-0 auto-rows-fr gap-2 overflow-hidden" aria-label="미리보기">
          {#each messageRows as row (row.role)}
            {@const renderedContent = messageForRole(previewMessages, row.role)}
            <article class="relative flex min-h-0 min-w-0 flex-col overflow-hidden rounded border border-border bg-card">
              <span
                class="pointer-events-none absolute -top-px right-[13px] rounded px-1.5 py-0.5 font-mono text-xs text-muted-foreground/70"
                >{row.role}</span
              >
              <div class="min-h-0 flex-1 overflow-auto p-3 pr-16">
                {#if previewError}
                  <p class="text-xs text-destructive">{previewError}</p>
                {:else if renderedContent.trim()}
                  <pre
                    class="whitespace-pre-wrap font-mono text-xs text-foreground">{#each renderedSegments(renderedContent) as segment, index (`${row.role}-${index}`)}<span
                        style={segment.kind === 'value' ? 'color: var(--color-content-added)' : undefined}>{segment.text}</span
                      >{/each}</pre>
                {:else}
                  <p class="text-xs text-muted-foreground">조건에 따라 이 메시지는 렌더링되지 않았습니다.</p>
                {/if}
              </div>
            </article>
          {/each}
        </section>
      {/if}
    </div>
  </div>
</div>

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
