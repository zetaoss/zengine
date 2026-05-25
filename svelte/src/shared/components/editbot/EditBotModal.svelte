<svelte:options runes={true} />

<script lang="ts">
  import { mdiArrowDown, mdiArrowUp, mdiAutoFix, mdiCreation, mdiDelete, mdiEye, mdiInformationOutline, mdiRefresh } from '@mdi/js'
  import { SvelteMap } from 'svelte/reactivity'

  import mwapi from '$lib/utils/mwapi'
  import { showToast } from '$shared/ui/toast/toast'
  import ZButton from '$shared/ui/ZButton.svelte'
  import ZButtonLink from '$shared/ui/ZButtonLink.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZModal from '$shared/ui/ZModal.svelte'
  import ZSelect from '$shared/ui/ZSelect.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import httpy from '$shared/utils/httpy'

  import { getPromptParts, renderFinalPrompt as renderFinal } from './renderer/promptRenderer'

  interface Target {
    title: string
    storeUrl: string
    requestType: 'create' | 'edit'
    pageId?: number
    existingContent?: string
  }

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

  type TemplateVariableMode = 'plain' | 'code' | 'list'

  interface TemplateVariable {
    name: string
    hashes: string
    mode: TemplateVariableMode
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
    show = false,
    target = null,
    onClose,
    onCreated,
  }: {
    show?: boolean
    target?: Target | null
    onClose?: () => void
    onCreated?: (data: StoreResp) => void
  } = $props()

  let promptItems = $state<PromptItem[]>([])
  let promptTitle = $state('')
  let promptListLoading = $state(false)
  let existingContentLoading = $state(false)
  let submitting = $state(false)
  let wasShown = false
  let existingContent = $state('')
  let originalDocumentContent = $state('')
  let displayTitle = $state('')
  let customFieldValues = $state<Record<string, string>>({})
  let listFieldRows = $state<Record<string, string[]>>({})
  let isRenderingEnabled = $state(true)
  let activeLabelInfo = $state<'page-title' | 'display-title' | null>(null)

  let filteredPromptItems = $derived(
    promptItems
      .filter((p) => p.request_type === target?.requestType)
      .sort((a, b) => {
        if (a.is_favorite !== b.is_favorite) return a.is_favorite ? -1 : 1
        return b.use_count - a.use_count
      }),
  )

  let currentPromptItem = $derived(filteredPromptItems.find((p) => p.title === promptTitle))
  let promptSelectItems = $derived<PromptSelectItem[]>(
    filteredPromptItems.length > 0
      ? filteredPromptItems.map((p) => ({ value: p.title, label: p.title }))
      : [{ value: '프롬프트 생성', label: '프롬프트 생성' }],
  )
  let llmPromptTemplate = $derived(currentPromptItem?.content ?? '')

  function clickOutside(node: HTMLElement, onOutsideClick: () => void) {
    let callback = onOutsideClick
    function handlePointerDown(event: PointerEvent) {
      if (node.contains(event.target as Node)) return
      callback()
    }
    document.addEventListener('pointerdown', handlePointerDown)
    return {
      update(nextOnOutsideClick: () => void) {
        callback = nextOnOutsideClick
      },
      destroy() {
        document.removeEventListener('pointerdown', handlePointerDown)
      },
    }
  }

  function parseTemplateVariableSpec(raw: string) {
    const [namePart, ...modeParts] = raw.split(':')
    const modeToken = (modeParts[0] ?? '').trim().toLowerCase()
    const mode: TemplateVariableMode = modeToken === 'code' || modeToken === 'list' ? modeToken : 'plain'
    return {
      name: namePart.trim(),
      mode,
    }
  }

  function isAutoTemplateVariable(name: string) {
    return name === '제목' || name === '기존 문서 내용'
  }

  function isDocumentContentVariable(name: string) {
    return name === '기존 문서 내용'
  }

  function resetDocumentContentField(name: string) {
    customFieldValues[name] = originalDocumentContent
  }

  function isDocumentContentDirty(name: string) {
    return customFieldValues[name] !== originalDocumentContent
  }

  function normalizeListRows(rows: string[]) {
    const values = rows.map((row) => row.trim()).filter((row) => row.length > 0)
    return [...values, '']
  }

  function setListRows(name: string, nextRows: string[]) {
    const normalized = normalizeListRows(nextRows)
    listFieldRows[name] = normalized
    customFieldValues[name] = normalized.filter((row) => row.length > 0).join('\n')
  }

  function updateListRow(name: string, index: number, value: string) {
    const rows = [...(listFieldRows[name] ?? [''])]
    rows[index] = value
    setListRows(name, rows)
  }

  function moveListRow(name: string, index: number, direction: -1 | 1) {
    const rows = (listFieldRows[name] ?? ['']).map((row) => row.trim()).filter((row) => row.length > 0)
    const nextIndex = index + direction
    if (index < 0 || index >= rows.length || nextIndex < 0 || nextIndex >= rows.length) return
    ;[rows[index], rows[nextIndex]] = [rows[nextIndex], rows[index]]
    setListRows(name, rows)
  }

  function removeListRow(name: string, index: number) {
    const rows = (listFieldRows[name] ?? ['']).map((row) => row.trim()).filter((row) => row.length > 0)
    if (index < 0 || index >= rows.length) return
    rows.splice(index, 1)
    setListRows(name, rows)
  }

  let templateVariables = $derived.by(() => {
    const vars = new SvelteMap<string, TemplateVariable>()
    // Detect patterns like "{# 추가정보}" or "{## 참고자료}"
    const matches = llmPromptTemplate.matchAll(/\{(#+\s*)([^}]+)\}/g)
    for (const match of matches) {
      const hashes = match[1]
      const parsed = parseTemplateVariableSpec(match[2].trim())
      if (!isAutoTemplateVariable(parsed.name) || (target?.requestType === 'edit' && isDocumentContentVariable(parsed.name))) {
        vars.set(parsed.name, { name: parsed.name, hashes, mode: parsed.mode })
      }
    }
    return Array.from(vars.values())
  })

  let editableTemplateVariables = $derived.by<TemplateVariable[]>(() => {
    if (target?.requestType !== 'edit') return templateVariables
    const rest = templateVariables.filter((v) => !isDocumentContentVariable(v.name))
    return [{ name: '기존 문서 내용', hashes: '# ', mode: 'plain' }, ...rest]
  })

  let previewParts = $derived(
    getPromptParts({
      template: llmPromptTemplate,
      displayTitle,
      existingContent,
      customFieldValues,
      rawMode: !isRenderingEnabled,
    }),
  )

  let renderedLlmInput = $derived.by(() => renderFinalPrompt())
  let modalTitle = $derived(`편집봇 - 문서 ${target?.requestType === 'edit' ? '편집' : '생성'}`)
  let modalTitleIconPath = $derived(target?.requestType === 'edit' ? mdiAutoFix : mdiCreation)
  let canSubmit = $derived(Boolean(target && renderedLlmInput.trim() && !existingContentLoading) && !submitting)

  $effect(() => {
    if (!show) {
      wasShown = false
      return
    }
    if (wasShown) return
    wasShown = true
    reset()
    if (target?.requestType === 'edit') {
      void fetchExistingContent(target.title)
    }
    void fetchPromptList()
  })

  $effect(() => {
    if (!show) return
    // Ensure all variables in the template have a corresponding field value
    for (const v of templateVariables) {
      if (!(v.name in customFieldValues)) {
        customFieldValues[v.name] = ''
      }
      if (v.mode === 'list' && !(v.name in listFieldRows)) {
        const seededRows = customFieldValues[v.name] ? customFieldValues[v.name].split('\n') : ['']
        setListRows(v.name, seededRows)
      }
    }
  })

  $effect(() => {
    if (!show || !filteredPromptItems.length) return
    // Auto-select first item if current promptTitle is not in the list
    if (!filteredPromptItems.some((p) => p.title === promptTitle)) {
      promptTitle = filteredPromptItems[0].title
    }
  })

  function reset() {
    promptTitle = target?.requestType === 'edit' ? '프롬프트 편집' : '프롬프트 생성'
    existingContentLoading = false
    existingContent = target?.existingContent ?? ''
    originalDocumentContent = ''
    displayTitle = (target?.title ?? '').replace(/\s*\(.*?\)$/, '').trim()
    customFieldValues = {}
    listFieldRows = {}
    isRenderingEnabled = true
    submitting = false
  }

  function asArray<T>(value: T[] | Record<string, T> | undefined): T[] {
    if (!value) return []
    return Array.isArray(value) ? value : Object.values(value)
  }

  function getRevisionContent(pageData: MwPage | undefined) {
    const revision = pageData?.revisions?.[0]
    return revision?.slots?.main?.content ?? revision?.content ?? ''
  }

  function renderFinalPrompt() {
    return renderFinal({
      template: llmPromptTemplate,
      displayTitle,
      existingContent,
      templateVariables,
      customFieldValues,
    })
  }

  function normalizeDisplayTitle(value: string) {
    return value.replace(/\s{2,}/g, ' ')
  }

  function shouldShowPageTitle(title: string | undefined, currentDisplayTitle: string) {
    if (!title) return false
    return title.trim() !== currentDisplayTitle.trim()
  }

  function resetDisplayTitleToPageTitle() {
    displayTitle = (target?.title ?? '').trim()
  }

  async function fetchPromptList() {
    promptListLoading = true
    try {
      const [data, err] = await httpy.get<PromptItem[]>('/api/editbot/prompts')
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

  async function fetchExistingContent(title: string) {
    existingContentLoading = true
    const [data, err] = await mwapi.get<MwRawTextResp>({
      action: 'query',
      prop: 'revisions',
      rvprop: 'content',
      rvslots: 'main',
      titles: title,
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
    originalDocumentContent = existingContent
    customFieldValues['기존 문서 내용'] = existingContent
  }

  async function ok() {
    if (!target || !canSubmit) return

    submitting = true
    const [data, err] = await httpy.post<StoreResp>(target.storeUrl, {
      page_id: target.pageId,
      prompt_title: promptTitle,
      request_type: target.requestType,
      llm_input: renderedLlmInput,
    })
    submitting = false

    if (err) {
      console.error(err)
      showToast(err.message || '편집봇 등록 실패')
      return
    }
    if (!data) return

    showToast(data.created ? '편집봇에 등록했습니다.' : '이미 편집봇에 등록되어 있습니다.', {
      timeout: 8000,
      placement: 'center',
      action: {
        label: '편집봇 작업목록 바로가기',
        href: '/tool/editbot',
      },
    })
    onCreated?.(data)
    onClose?.()
  }

  function cancel() {
    onClose?.()
  }
</script>

<ZModal
  {show}
  title={modalTitle}
  titleIconPath={modalTitleIconPath}
  titleIconAtEnd={true}
  okText={submitting ? '등록 중' : '등록'}
  okColor="primary"
  okDisabled={!canSubmit}
  backdropClosable={true}
  panelClass="h-[80vh] w-[90vw] md:w-[75vw]"
  on:ok={ok}
  on:cancel={cancel}
>
  <div class="flex h-full min-h-0 flex-col gap-4 overflow-y-auto">
    <div class="flex flex-col gap-3 border-b border-(--border-color-subtle) pb-4">
      <div class="flex items-center gap-2">
        <span class="w-20 shrink-0 text-sm font-semibold text-(--color-subtle)">템플릿 선택</span>
        {#if promptListLoading}
          <div class="flex h-9 items-center">
            <ZSpinner />
          </div>
        {:else}
          <ZSelect
            bind:value={promptTitle}
            items={promptSelectItems}
            class="min-w-0 flex-1"
          >
            {#snippet item(p: PromptSelectItem)}
              {@const promptMeta = filteredPromptItems.find((item) => item.title === p.value)}
              <div class="flex flex-1 items-center justify-between gap-2 overflow-hidden">
                <div class="truncate">{p.label}</div>
                <div class="w-20 shrink-0 text-right text-xs opacity-50">
                  {#if promptMeta?.is_favorite}⭐ {/if}{promptMeta?.use_count ?? 0} runs
                </div>
              </div>
            {/snippet}
          </ZSelect>
          <div class="flex shrink-0 items-center gap-1">
            {#if currentPromptItem}
              <ZButtonLink
                color="default"
                href={`/tool/editbot/prompts/${currentPromptItem.id}`}
                target="_blank"
                title={`${promptTitle} 편집`}
              >
                <ZIcon path={mdiEye} />
              </ZButtonLink>
            {/if}
            <ZButton color="default" title="새로고침" onclick={() => void fetchPromptList()}>
              <ZIcon path={mdiRefresh} />
            </ZButton>
          </div>
        {/if}
      </div>

      {#if shouldShowPageTitle(target?.title, displayTitle)}
        <div class="flex items-center gap-2">
          <div
            class="relative flex w-24 shrink-0 items-center gap-1 text-sm font-semibold text-(--color-subtle)"
            use:clickOutside={() => {
              if (activeLabelInfo === 'page-title') activeLabelInfo = null
            }}
          >
            <span>표제어</span>
            <button
              type="button"
              class="inline-flex opacity-70 hover:opacity-100"
              aria-label="page title 도움말"
              aria-expanded={activeLabelInfo === 'page-title'}
              onclick={() => (activeLabelInfo = activeLabelInfo === 'page-title' ? null : 'page-title')}
            >
              <ZIcon path={mdiInformationOutline} />
            </button>
            {#if activeLabelInfo === 'page-title'}
              <div class="absolute top-full left-0 z-10 mt-1 rounded border border-(--border-color-subtle) bg-(--background-color-base) px-2 py-1 text-xs font-normal whitespace-nowrap shadow">
                표제어 - 표제어(page title, 페이지 제목)는 문서를 식별하는 고유한 이름이며, URL과 링크의 기준이 됩니다.
              </div>
            {/if}
          </div>
          <input type="text" class="z-input flex-1 bg-(--background-color-neutral-subtle) opacity-80" value={target?.title} readonly />
        </div>
      {/if}

      <div class="flex items-center gap-2">
        <div
          class="relative flex w-24 shrink-0 items-center gap-1 text-sm font-semibold text-(--color-subtle)"
          use:clickOutside={() => {
            if (activeLabelInfo === 'display-title') activeLabelInfo = null
          }}
        >
          <span>제목</span>
          {#if shouldShowPageTitle(target?.title, displayTitle)}
            <button
              type="button"
              class="inline-flex opacity-70 hover:opacity-100"
              aria-label="display title 도움말"
              aria-expanded={activeLabelInfo === 'display-title'}
              onclick={() => (activeLabelInfo = activeLabelInfo === 'display-title' ? null : 'display-title')}
            >
              <ZIcon path={mdiInformationOutline} />
            </button>
            {#if activeLabelInfo === 'display-title'}
              <div class="absolute top-full left-0 z-10 mt-1 rounded border border-(--border-color-subtle) bg-(--background-color-base) px-2 py-1 text-xs font-normal whitespace-nowrap shadow">
                제목(display title, 표시 제목)은 문서가 설명하는 대상을 자연스럽게 지칭하기 위한 이름입니다.
              </div>
            {/if}
          {/if}
        </div>
        <input
          type="text"
          class="z-input flex-1"
          bind:value={displayTitle}
          oninput={(e) => (displayTitle = normalizeDisplayTitle((e.currentTarget as HTMLInputElement).value))}
          onblur={() => (displayTitle = normalizeDisplayTitle(displayTitle).trim())}
          placeholder="문서 내에서 사용될 제목"
        />
        {#if shouldShowPageTitle(target?.title, displayTitle)}
          <ZButton color="default" size="small" title="표제어로 되돌리기" onclick={resetDisplayTitleToPageTitle}>
            리셋
          </ZButton>
        {/if}
      </div>
    </div>

    <div class="grid min-h-0 flex-1 grid-cols-1 gap-4 overflow-hidden md:grid-cols-2">
      <div class="flex min-h-0 flex-col gap-4 overflow-y-auto pr-1">
        {#each editableTemplateVariables as v (v.name)}
          <div class="flex flex-col gap-1">
            <div class="flex items-center justify-between gap-2 text-sm text-(--color-subtle)">
              <label for={`editbot-field-${v.name}`} class="font-semibold">{v.name}</label>
              {#if isDocumentContentVariable(v.name) && isDocumentContentDirty(v.name)}
                <ZButton color="default" size="small" title="문서내용 리셋" onclick={() => resetDocumentContentField(v.name)}>
                  리셋
                </ZButton>
              {/if}
            </div>
            {#if v.mode === 'list'}
              <div class="flex flex-col gap-1">
                {#each listFieldRows[v.name] ?? [''] as rowValue, rowIndex (rowIndex)}
                  <div class="flex items-center gap-1">
                    <input
                      id={rowIndex === 0 ? `editbot-field-${v.name}` : undefined}
                      type="text"
                      class="z-input min-h-0 flex-1 text-xs"
                      value={rowValue}
                      oninput={(e) => updateListRow(v.name, rowIndex, (e.currentTarget as HTMLInputElement).value)}
                      placeholder={`${v.name} 항목`}
                    />
                    <ZButton color="default" size="small" title="위로 이동" disabled={rowIndex === 0 || !rowValue.trim()} onclick={() => moveListRow(v.name, rowIndex, -1)}>
                      <ZIcon path={mdiArrowUp} />
                    </ZButton>
                    <ZButton
                      color="default"
                      size="small"
                      title="아래로 이동"
                      disabled={rowIndex >= (listFieldRows[v.name]?.length ?? 1) - 2 || !rowValue.trim()}
                      onclick={() => moveListRow(v.name, rowIndex, 1)}
                    >
                      <ZIcon path={mdiArrowDown} />
                    </ZButton>
                    <ZButton color="default" size="small" title="행 삭제" disabled={!rowValue.trim()} onclick={() => removeListRow(v.name, rowIndex)}>
                      <ZIcon path={mdiDelete} />
                    </ZButton>
                  </div>
                {/each}
              </div>
            {:else}
              <textarea
                id={`editbot-field-${v.name}`}
                class="z-input h-32 min-h-0 resize-none font-mono text-xs"
                bind:value={customFieldValues[v.name]}
                placeholder={`${v.name} 내용을 입력하세요.`}
              ></textarea>
            {/if}
          </div>
        {:else}
          <div class="flex h-32 items-center justify-center rounded border border-dashed border-(--border-color-subtle) text-sm text-(--color-subtle)">
            입력할 추가 항목이 없습니다.
          </div>
        {/each}
      </div>

      <div class="flex min-h-0 flex-col gap-1">
        <div class="flex items-center justify-between gap-2 text-sm text-(--color-subtle)">
          <label for="editbot-llm-input" class="font-semibold">프롬프트</label>
          <div class="flex overflow-hidden rounded border border-(--border-color-subtle)">
            <button
              type="button"
              class="px-2.5 py-0.5 text-xs transition-colors {isRenderingEnabled ? 'bg-transparent' : 'bg-(--background-color-neutral-subtle) font-semibold'}"
              onclick={() => (isRenderingEnabled = false)}
            >
              템플릿
            </button>
            <button
              type="button"
              class="border-l border-(--border-color-subtle) px-2.5 py-0.5 text-xs transition-colors {!isRenderingEnabled ? 'bg-transparent' : 'bg-(--background-color-neutral-subtle) font-semibold'}"
              onclick={() => (isRenderingEnabled = true)}
            >
              렌더링
            </button>
          </div>
        </div>
        <div
          class="z-input h-full min-h-0 overflow-y-auto whitespace-pre-wrap bg-(--background-color-neutral-subtle) p-2 font-mono text-xs opacity-90"
        >
          {#each previewParts as part, i (i)}
            {#if part.type === 'inline'}
              <span class={!isRenderingEnabled ? 'inline-var' : 'inline-var inline-var--no-bg'}>{part.text}</span>
            {:else if part.type === 'block'}
              <span class={!isRenderingEnabled ? 'block-var' : 'block-var block-var--no-bg'}>{part.text}</span>
            {:else if part.type === 'code'}
              <span class="code-block">{part.text}</span>
            {:else}
              {part.text}
            {/if}
          {/each}
          {#if previewParts.length === 0}
            <span class="text-(--color-subtle)">템플릿이 로딩되면 렌더링된 프롬프트가 표시됩니다.</span>
          {/if}
        </div>
      </div>
    </div>
  </div>
</ZModal>

<style>
  .inline-var {
    color: var(--blue-9);
    background-color: var(--blue-3);
    padding: 0 2px;
    border-radius: 4px;
    border: 1px solid var(--blue-6);
    font-weight: 600;
  }

  .block-var {
    color: var(--brown-9);
    background-color: var(--brown-3);
    padding: 0 2px;
    border: 1px solid var(--brown-6);
    border-radius: 4px;
    font-weight: 600;
  }

  .inline-var--no-bg {
    background-color: transparent;
    border-color: transparent;
    padding: 0;
  }

  .block-var--no-bg {
    background-color: transparent;
    border-color: transparent;
  }

  .code-block {
    color: var(--slate-11);
  }
</style>
