<svelte:options runes={true} />

<script lang="ts">
  import { mdiEye, mdiHelpCircleOutline, mdiRefresh } from '@mdi/js'
  import { SvelteMap } from 'svelte/reactivity'

  import mwapi from '$lib/utils/mwapi'
  import CButton from '$shared/ui/CButton.svelte'
  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSelect from '$shared/ui/ZSelect.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import ZTooltip from '$shared/ui/Ztooltip.svelte'
  import httpy from '$shared/utils/httpy'

  import { getPromptParts, renderFinalPrompt } from './promptRenderer'

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

  type TemplateVariableMode = 'plain' | 'code'

  interface TemplateVariable {
    name: string
    hashes: string
    mode: TemplateVariableMode
    blockLang: string
  }

  type PreviewPartType = 'plain' | 'inline' | 'block' | 'code' | 'text'

  interface PreviewPart {
    text: string
    type: PreviewPartType
    source?: 'preset' | 'custom'
    ordinal?: number
    label?: string
  }

  interface MwRevision {
    content?: string
    slots?: {
      main?: {
        content?: string
      }
    }
  }

  interface MwPage {
    title: string
    missing?: boolean
    revisions?: MwRevision[]
  }

  interface MwRawTextResp {
    query?: {
      pages?: MwPage[] | Record<string, MwPage>
    }
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
    requestType: 'create' | 'edit'
    pageId: number | undefined
    title: string
  } = $props()

  let promptItems = $state<PromptItem[]>([])
  let promptTitle = $state('')
  let promptListLoading = $state(false)
  let existingContentLoading = $state(false)
  let submitting = $state(false)
  let existingContent = $state('')
  let displayTitle = $state('')
  let notesEnabled = $state(false)
  let notes = $state('')
  let customFieldValues = $state<Record<string, string>>({})
  let lastPromptTitle = $state('')

  let filteredPromptItems = $derived(
    promptItems
      .filter((p) => p.request_type === requestType)
      .sort((a, b) => {
        if (a.is_favorite !== b.is_favorite) return a.is_favorite ? -1 : 1
        return b.use_count - a.use_count
      }),
  )

  let currentPromptItem = $derived(filteredPromptItems.find((p) => p.title === promptTitle))
  let promptSelectItems = $derived<PromptSelectItem[]>(
    filteredPromptItems.length > 0
      ? filteredPromptItems.map((p) => ({ value: p.title, label: p.title }))
      : [{ value: '', label: '프롬프트 없음' }],
  )

  let llmPromptTemplate = $derived(filteredPromptItems.find((p) => p.title === promptTitle)?.content ?? '')

  let effectiveNotes = $derived(notesEnabled ? notes : '')

  let templateVariables = $derived.by(() => {
    const vars = new SvelteMap<string, TemplateVariable>()
    const matches = llmPromptTemplate.matchAll(/\{(#+\s*)([^}]+)\}/g)
    for (const match of matches) {
      const hashes = match[1]
      const [namePart, ...modeParts] = match[2].split(':')
      const modeToken = (modeParts[0] ?? '').trim().toLowerCase()
      const mode: TemplateVariableMode = modeToken === 'code' ? 'code' : 'plain'
      const name = namePart.trim()
      if (name === '제목' || name === '기존 문서 내용') continue
      vars.set(name, { name, hashes, mode, blockLang: (modeParts[1] ?? '').trim() })
    }
    return Array.from(vars.values())
  })

  let renderedLlmInput = $derived(
    renderFinalPrompt({
      template: llmPromptTemplate,
      displayTitle,
      existingContent,
      customFieldValues,
      notes: effectiveNotes,
    }),
  )

  let canSubmit = $derived(Boolean(promptTitle && renderedLlmInput.trim() && !existingContentLoading) && !submitting)
  let previewParts = $derived<PreviewPart[]>(
    getPromptParts({
      template: llmPromptTemplate,
      displayTitle,
      existingContent,
      customFieldValues,
      notes: effectiveNotes,
      rawMode: false,
      preserveEmptyBlockPlaceholders: true,
    }),
  )

  $effect(() => {
    displayTitle = (title ?? '').replace(/\s*\(.*?\)$/, '').trim()
    existingContent = ''
    notesEnabled = false
    notes = ''
    customFieldValues = {}
    lastPromptTitle = ''
    void fetchPromptList()
    if (requestType === 'edit') {
      void fetchExistingContent(title)
    }
  })

  $effect(() => {
    if (!filteredPromptItems.length) return
    if (!filteredPromptItems.some((p) => p.title === promptTitle)) {
      promptTitle = filteredPromptItems[0].title
    }
  })

  $effect(() => {
    if (!promptTitle || promptTitle === lastPromptTitle) return
    lastPromptTitle = promptTitle
    const nextValues: Record<string, string> = {}
    for (const variable of templateVariables) {
      nextValues[variable.name] = ''
    }
    customFieldValues = nextValues
  })

  function asArray<T>(value: T[] | Record<string, T> | undefined): T[] {
    if (!value) return []
    return Array.isArray(value) ? value : Object.values(value)
  }

  function getRevisionContent(pageData: MwPage | undefined) {
    const revision = pageData?.revisions?.[0]
    return revision?.slots?.main?.content ?? revision?.content ?? ''
  }

  async function fetchPromptList() {
    promptListLoading = true
    try {
      const [data, err] = await httpy.get<PromptItem[]>('/api/ai-edit/prompts')
      if (err) {
        console.error(err)
        return
      }
      promptItems = data ?? []
    } catch (e) {
      console.error(e)
    } finally {
      promptListLoading = false
    }
  }

  async function refreshPromptList() {
    await fetchPromptList()
  }

  function updateFieldValue(name: string, value: string) {
    customFieldValues = {
      ...customFieldValues,
      [name]: value,
    }
  }

  async function fetchExistingContent(pageTitle: string) {
    existingContentLoading = true
    const [data, err] = await mwapi.get<MwRawTextResp>({
      action: 'query',
      prop: 'revisions',
      rvprop: 'content',
      rvslots: 'main',
      titles: pageTitle,
    })
    existingContentLoading = false

    if (err) {
      console.error(err)
      showToast(err.message || '기존 문서를 불러오지 못했습니다.')
      return
    }

    const pageData = asArray<MwPage>(data?.query?.pages)[0]
    if (!pageData || pageData.missing) {
      showToast('기존 문서를 찾을 수 없습니다.')
      return
    }

    existingContent = getRevisionContent(pageData)
  }

  function releaseEditWarning() {
    window.jQuery?.(window).off('beforeunload.editwarning')
  }

  async function submit() {
    if (!canSubmit) return

    submitting = true
    const [data, err] = await httpy.post<StoreResp>('/api/ai-edit', {
      page_id: pageId,
      title,
      prompt_title: promptTitle,
      request_type: requestType,
      llm_input: renderedLlmInput,
    })
    submitting = false

    if (err) {
      console.error(err)
      return
    }
    if (!data) return

    releaseEditWarning()
    window.location.assign(`/tool/ai-edit/tasks/${data.id}`)
  }

  async function handleSubmitClick() {
    if (!canSubmit) return
    await submit()
  }
</script>

<div class="flex h-full min-h-0 flex-col gap-3 border p-4">
  <div class="flex items-center gap-2">
    <span class="w-20 shrink-0 text-right text-sm font-semibold text-slate-700">템플릿</span>
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

  <div class="flex flex-wrap items-center gap-2">
    <span class="w-20 shrink-0 text-right text-sm font-semibold text-slate-700">제목</span>
    <ZTooltip content="제목은 AI 편집에 전달되는 표시용 제목입니다. 원본 표제어와 다를 수 있습니다." ariaLabel="제목 도움말">
      <ZIcon path={mdiHelpCircleOutline} class="h-5 w-5 text-slate-500" aria-hidden="true" />
    </ZTooltip>
    <input type="text" class="min-w-0 flex-1 rounded border px-2 py-1 text-sm" bind:value={displayTitle} />
    <span class="w-20 shrink-0 text-right text-sm font-semibold text-slate-700">표제어</span>
    <input type="text" class="min-w-0 flex-1 rounded border bg-slate-100 px-2 py-1 text-sm text-slate-600" value={title} readonly />
  </div>

  {#if templateVariables.length > 0}
    {#each templateVariables as variable (variable.name)}
      <div class="flex flex-col gap-1">
        <div class="flex items-center justify-between gap-2 text-sm text-slate-700">
          <label class="flex items-center gap-2 font-semibold" for={`ai-edit-panel-field-${variable.name}`}>
            <span>{variable.name}</span>
            {#if variable.mode === 'code'}
              <span class="inline-flex items-center rounded border border-slate-200 px-1 py-0.5 text-[10px] font-medium leading-none text-slate-500 select-none">
                code
              </span>
            {/if}
          </label>
        </div>
        <textarea
          id={`ai-edit-panel-field-${variable.name}`}
          class={`min-h-24 rounded border border-slate-300 px-2 py-1 text-sm ${variable.mode === 'code' ? 'font-mono' : ''}`}
          value={customFieldValues[variable.name] ?? ''}
          oninput={(event) => updateFieldValue(variable.name, (event.currentTarget as HTMLTextAreaElement).value)}
          placeholder={`${variable.name} 내용을 입력하세요.`}
        ></textarea>
      </div>
    {/each}
  {/if}

  <div class="flex items-center gap-2 text-sm text-slate-700">
    <label for="ai-edit-panel-notes-enabled" class="font-semibold">비고</label>
    <label class="flex items-center gap-1 text-xs text-slate-500" for="ai-edit-panel-notes-enabled">
      <input id="ai-edit-panel-notes-enabled" type="checkbox" bind:checked={notesEnabled} class="h-3.5 w-3.5 rounded border-slate-300" />
      <span>사용</span>
    </label>
  </div>
  {#if notesEnabled}
    <textarea
      id="ai-edit-panel-additional-content"
      class="min-h-24 rounded border border-slate-300 px-2 py-1 text-sm"
      bind:value={notes}
      placeholder="프롬프트에 덧붙일 내용을 입력하세요."
    ></textarea>
  {/if}

  <hr class="border-slate-200" />

  <div class="flex min-h-0 flex-1 flex-col gap-1">
    <div class="flex items-center justify-between gap-2 text-sm text-slate-700">
      <label for="ai-edit-panel-prompt-preview" class="font-semibold">프롬프트</label>
    </div>
    <div
      id="ai-edit-panel-prompt-preview"
      class="z-input min-h-0 flex-1 overflow-y-auto whitespace-pre-wrap rounded border border-slate-300 bg-slate-50 p-2 font-mono text-xs text-slate-700"
    >
      {#each previewParts as part, i (i)}
        {@const partType = part.type as PreviewPartType}
        {@const isEmptyPart = part.text === ''}
        {#if partType === 'text'}
          <div class="rounded border">{part.text}</div>
        {:else}
          {#if partType === 'plain'}
            {part.text}
          {:else if partType === 'inline'}
            {#if isEmptyPart}
              <span
                class="inline-flex items-center justify-center rounded border border-dashed border-slate-300 px-2 text-slate-400 select-none"
              >
                {part.label ?? ''}
              </span>
            {:else}
              <span class="inline-flex items-stretch overflow-hidden rounded border border-slate-300 bg-slate-100 text-slate-700">
                <span class="px-1">{part.text}</span>
              </span>
            {/if}
          {:else if isEmptyPart}
            <div
              class="flex w-full items-center justify-center rounded border border-dashed border-slate-300 px-2 text-slate-400 select-none"
            >
              {part.label ?? ''}
            </div>
          {:else}
            <div
              class="flex w-full min-h-2 items-stretch overflow-hidden rounded border border-slate-300 bg-slate-100 text-slate-700"
            >
              <span class="px-1 whitespace-pre-wrap">{part.text}</span>
            </div>
          {/if}
        {/if}
      {/each}
    </div>
  </div>

  <div class="flex justify-center">
    <CButton type="button" variant="default" disabled={!canSubmit} onclick={() => void handleSubmitClick()}>
      {submitting ? '등록 중' : 'AI 편집 등록'}
    </CButton>
  </div>
</div>
