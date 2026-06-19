<svelte:options runes={true} />

<script lang="ts">
  import { mdiRefresh } from '@mdi/js'
  import { onDestroy, tick } from 'svelte'

  import ZIcon from '$shared/ui/ZIcon.svelte'

  import { copyEditorSelection } from './editor/editorCopy'
  import { getContentEditableText, getNodeElement, getTextOffset, setCaret, setCaretAfter, setCaretInText } from './editor/editorDom'
  import { isReservedAutoVariableName, type PromptPart } from './prompt/promptRenderer'

  let {
    template = $bindable(''),
    parts,
    displayTitle = $bindable(''),
    variableValues = $bindable({}),
    editable = false,
    reservedDefaults = {},
    isExternalContent = false,
    externalContentDirty = false,
    id = 'ai-edit-prompt-preview',
    class: className = '',
    emptyText = '프롬프트 내용이 없습니다.',
    refreshKey = 0,
    oninput = undefined,
    onTitleReset = undefined,
    onContentResetRequest = undefined,
  }: {
    template?: string
    parts: PromptPart[]
    displayTitle?: string
    variableValues?: Record<string, string>
    editable?: boolean
    reservedDefaults?: Record<string, string>
    isExternalContent?: boolean
    externalContentDirty?: boolean
    id?: string
    class?: string
    emptyText?: string
    refreshKey?: number
    oninput?: () => void
    onTitleReset?: () => void
    onContentResetRequest?: () => void
  } = $props()

  type CaretSnapshot = { kind: 'template'; offset: number } | { kind: 'variable'; label: string; offset: number }
  type ResetUndoEntry = { label: string; value: string }

  let editorEl = $state<HTMLDivElement | undefined>()
  let contentEl = $state<HTMLDivElement | undefined>()
  let liveParts = $state<PromptPart[]>([])
  let isComposing = false
  let commitTimer: ReturnType<typeof setTimeout> | undefined
  let editorActionSerial = 0
  let lastRefreshKey = $state<number | undefined>(undefined)
  let focusedVariableLabel = $state<string | undefined>(undefined)
  let resetUndoStack: ResetUndoEntry[] = []

  $effect(() => {
    if (isEditorActive()) return
    liveParts = parts
  })

  $effect(() => {
    if (refreshKey === lastRefreshKey) return
    lastRefreshKey = refreshKey
    void refreshRenderedParts()
  })

  onDestroy(() => {
    if (commitTimer !== undefined) clearTimeout(commitTimer)
  })

  function handleInput() {
    oninput?.()
  }

  function handleEditorInput(event: Event) {
    if (isComposing || (event instanceof InputEvent && event.isComposing)) return
    resetUndoStack = []
    const shouldRestoreCaret =
      !(event instanceof InputEvent) || (event.inputType !== 'insertParagraph' && event.inputType !== 'insertLineBreak')
    scheduleEditorCommit({ restoreCaret: shouldRestoreCaret })
  }

  function handleCompositionStart() {
    isComposing = true
  }

  function handleCompositionEnd() {
    isComposing = false
    scheduleEditorCommit()
  }

  function handleEditorFocusOut() {
    const actionSerial = editorActionSerial
    setTimeout(() => {
      if (actionSerial !== editorActionSerial) return
      if (isEditorActive()) return
      void commitEditorChange({ refreshParts: true })
    }, 0)
  }

  function scheduleEditorCommit({ restoreCaret = true }: { restoreCaret?: boolean } = {}) {
    if (commitTimer !== undefined) clearTimeout(commitTimer)
    commitTimer = setTimeout(() => {
      commitTimer = undefined
      void commitEditorChange({ restoreCaret })
    }, 0)
  }

  async function commitEditorChange({
    refreshParts = false,
    restoreCaret: shouldRestoreCaret = true,
  }: { refreshParts?: boolean; restoreCaret?: boolean } = {}) {
    if (!contentEl) return
    const caret = shouldRestoreCaret ? getCaretSnapshot() : undefined
    syncEditorState(contentEl)
    handleInput()
    await tick()
    if (refreshParts) liveParts = parts
    if (!refreshParts && shouldRestoreCaret) restoreCaret(caret)
  }

  function handleEditorPaste(event: ClipboardEvent) {
    const text = event.clipboardData?.getData('text/plain')
    if (text === undefined) return

    event.preventDefault()
    document.execCommand('insertText', false, text)
  }

  function handleEditorCopy(event: ClipboardEvent) {
    if (!contentEl) return

    copyEditorSelection(contentEl, event)
  }

  function handleEditorKeydown(event: KeyboardEvent) {
    if ((event.ctrlKey || event.metaKey) && !event.shiftKey && event.key.toLowerCase() === 'z' && resetUndoStack.length > 0) {
      event.preventDefault()
      void undoLastReset()
    }
  }

  function handleControlPointerDown(event: MouseEvent) {
    event.preventDefault()
  }

  function handleResetClick(event: MouseEvent, part: PromptPart) {
    event.stopPropagation()
    handlePartReset(part)
  }

  function handleVariableHeaderKeydown(event: KeyboardEvent, label: string | undefined) {
    if (event.key !== 'Enter' && event.key !== ' ') return

    event.preventDefault()
    void focusVariableValue(label)
  }

  function prepareResetAction() {
    editorActionSerial += 1
    if (commitTimer !== undefined) {
      clearTimeout(commitTimer)
      commitTimer = undefined
    }
  }

  function handleReset(label: string) {
    const previousValue = getVariableValue(label)

    if (label === '제목' && onTitleReset) {
      prepareResetAction()
      pushResetUndo(label, previousValue)
      onTitleReset()
      return
    }

    if (label === '기존 문서 내용' && onContentResetRequest) {
      prepareResetAction()
      pushResetUndo(label, previousValue)
      onContentResetRequest()
      return
    }

    const isReserved = isReservedAutoVariableName(label)
    if (isExternalContent && label === '기존 문서 내용') return
    const defaultValue = isReserved ? reservedDefaults[label] : ''
    if (defaultValue === undefined) return

    prepareResetAction()
    pushResetUndo(label, previousValue)
    setVariableValue(label, defaultValue)
    handleInput()
    void refreshPartsAfterStateChange()
  }

  function handlePartReset(part: PromptPart) {
    if (!part.label) return
    handleReset(part.label)
  }

  async function refreshPartsAfterStateChange() {
    await tick()
    liveParts = parts
    await tick()
    syncRenderedVariableValues()
  }

  function pushResetUndo(label: string, value: string) {
    resetUndoStack = [...resetUndoStack, { label, value }]
  }

  async function undoLastReset() {
    const entry = resetUndoStack.at(-1)
    if (!entry) return

    resetUndoStack = resetUndoStack.slice(0, -1)
    prepareResetAction()
    setVariableValue(entry.label, entry.value)
    handleInput()
    await refreshPartsAfterStateChange()
    await focusVariableValue(entry.label)
  }

  async function refreshRenderedParts() {
    liveParts = parts
    await tick()
    syncRenderedVariableValues()
  }

  function syncRenderedVariableValues() {
    if (!contentEl) return

    const valueByLabel = new Map(
      liveParts
        .filter((part): part is PromptPart & { label: string } => typeof part.label === 'string')
        .map((part) => [part.label, part.text]),
    )

    for (const variableEl of Array.from(contentEl.querySelectorAll<HTMLElement>('[data-variable-label]'))) {
      const label = variableEl.dataset.variableLabel
      const value = label ? valueByLabel.get(label) : undefined
      const valueEl = variableEl.querySelector<HTMLElement>('[data-variable-value]')
      if (value === undefined || !valueEl) continue
      setVariableBlockState(variableEl, !value.trim(), variableEl.dataset.valueFocused === 'true')
      if (valueEl.textContent !== value) valueEl.textContent = value
    }
  }

  function isEditorActive() {
    return editorEl ? editorEl.contains(document.activeElement) : false
  }

  function syncEditorState(root: HTMLElement) {
    const nextTemplate = serializeTemplate(root.childNodes)
    if (nextTemplate !== template) template = nextTemplate
    if (editorEl) syncVariableValues(editorEl)
  }

  function serializeTemplate(nodes: NodeListOf<ChildNode>) {
    let value = ''

    for (const node of Array.from(nodes)) {
      if (node.nodeType === Node.TEXT_NODE) {
        const text = node.textContent ?? ''
        if (text.trim()) value += text
        continue
      }

      if (!(node instanceof HTMLElement)) continue

      if (node.dataset.editorPlaceholder) continue

      if (node.dataset.templateText) {
        value += getSerializedTemplateText(node)
        continue
      }

      const variableLabel = node.dataset.variableLabel
      if (variableLabel) {
        value += `{${variableLabel}}`
        continue
      }

      if (node.tagName === 'BR') {
        value += '\n'
        continue
      }

      value += serializeTemplate(node.childNodes)
    }

    return value
  }

  function syncVariableValues(root: HTMLElement) {
    const variableEls = Array.from(root.querySelectorAll<HTMLElement>('[data-variable-label]'))
    for (const variableEl of variableEls) {
      const label = variableEl.dataset.variableLabel
      const valueEl = variableEl.querySelector<HTMLElement>('[data-variable-value]')
      if (!label || !valueEl || !isVariableEditable(label)) continue

      const value = getContentEditableText(valueEl)
      setVariableBlockState(variableEl, !value.trim(), variableEl.dataset.valueFocused === 'true')
      setVariableValue(label, value)
    }
  }

  function setVariableBlockState(variableEl: HTMLElement, isEmpty: boolean, isFocused: boolean) {
    variableEl.dataset.emptyValue = isEmpty ? 'true' : 'false'
    variableEl.dataset.valueFocused = isFocused ? 'true' : 'false'
    const headerEl = variableEl.querySelector<HTMLElement>('[data-variable-header]')
    if (!headerEl) return

    const isCollapsed = isEmpty && !isFocused
    headerEl.classList.toggle('text-a-slate-400', isCollapsed)
    headerEl.classList.toggle('text-a-slate-700', !isCollapsed)
    headerEl.classList.toggle('select-none', isCollapsed)
  }

  async function focusVariableValue(label: string | undefined) {
    if (!contentEl || !label || !isVariableEditable(label)) return

    focusedVariableLabel = label
    await tick()
    const variableEl = findVariableElement(label)
    const valueEl = variableEl?.querySelector<HTMLElement>('[data-variable-value]')
    if (!variableEl || !valueEl) return

    setVariableBlockState(variableEl, !getContentEditableText(valueEl).trim(), true)
    valueEl.focus()
    setCaretInText(valueEl, getContentEditableText(valueEl).length)
  }

  function handleVariableValueFocus(part: PromptPart) {
    focusedVariableLabel = part.label
    const variableEl = part.label ? findVariableElement(part.label) : undefined
    const valueEl = variableEl?.querySelector<HTMLElement>('[data-variable-value]')
    if (!variableEl || !valueEl) return

    setVariableBlockState(variableEl, !getContentEditableText(valueEl).trim(), true)
  }

  function handleVariableValueBlur(part: PromptPart) {
    if (focusedVariableLabel === part.label) focusedVariableLabel = undefined
    const variableEl = part.label ? findVariableElement(part.label) : undefined
    const valueEl = variableEl?.querySelector<HTMLElement>('[data-variable-value]')
    if (!variableEl || !valueEl) return

    setVariableBlockState(variableEl, !getContentEditableText(valueEl).trim(), false)
  }

  function findVariableElement(label: string) {
    return Array.from(contentEl?.querySelectorAll<HTMLElement>('[data-variable-label]') ?? []).find(
      (element) => element.dataset.variableLabel === label,
    )
  }

  function getCaretSnapshot(): CaretSnapshot | undefined {
    if (!contentEl) return undefined
    const selection = window.getSelection()
    const focusNode = selection?.focusNode
    if (!selection || !focusNode || !contentEl.contains(focusNode)) return undefined

    const focusElement = getNodeElement(focusNode)
    const valueEl = focusElement?.closest<HTMLElement>('[data-variable-value]')
    const variableEl = valueEl?.closest<HTMLElement>('[data-variable-label]')
    const label = variableEl?.dataset.variableLabel
    if (valueEl && label) {
      return {
        kind: 'variable',
        label,
        offset: getTextOffset(valueEl, focusNode, selection.focusOffset),
      }
    }

    return {
      kind: 'template',
      offset: getTemplateOffset(contentEl.childNodes, focusNode, selection.focusOffset),
    }
  }

  function getTemplateOffset(nodes: NodeListOf<ChildNode>, focusNode: Node, focusOffset: number): number {
    let offset = 0

    for (const node of Array.from(nodes)) {
      if (node === focusNode) {
        if (node.nodeType === Node.TEXT_NODE) return offset + focusOffset
        return offset
      }

      if (node.nodeType === Node.TEXT_NODE) {
        const text = node.textContent ?? ''
        if (text.trim()) offset += text.length
        continue
      }

      if (!(node instanceof HTMLElement)) continue
      if (node.contains(focusNode)) {
        if (node.dataset.templateText) {
          return offset + getTemplatePrefix(node).length + getTextOffset(node, focusNode, focusOffset)
        }
        return offset + getTemplateOffset(node.childNodes, focusNode, focusOffset)
      }

      if (node.dataset.editorPlaceholder) continue
      if (node.dataset.templateText) {
        offset += getSerializedTemplateText(node).length
        continue
      }

      const variableLabel = node.dataset.variableLabel
      if (variableLabel) {
        offset += `{${variableLabel}}`.length
        continue
      }

      if (node.tagName === 'BR') {
        offset += 1
        continue
      }

      offset += serializeTemplate(node.childNodes).length
    }

    return offset
  }

  function restoreCaret(snapshot: CaretSnapshot | undefined) {
    if (!contentEl || !snapshot) return

    if (snapshot.kind === 'variable') {
      const variableEl = Array.from(contentEl.querySelectorAll<HTMLElement>('[data-variable-label]')).find(
        (element) => element.dataset.variableLabel === snapshot.label,
      )
      const valueEl = variableEl?.querySelector<HTMLElement>('[data-variable-value]')
      if (valueEl) {
        setCaretInText(valueEl, snapshot.offset)
        return
      }
    }

    setTemplateCaret(contentEl.childNodes, snapshot.offset)
  }

  function setTemplateCaret(nodes: NodeListOf<ChildNode>, targetOffset: number) {
    let remaining = targetOffset
    let lastElement: HTMLElement | undefined

    for (const node of Array.from(nodes)) {
      if (node.nodeType === Node.TEXT_NODE) {
        const textLength = node.textContent?.length ?? 0
        if (remaining <= textLength) {
          setCaret(node, remaining)
          return
        }
        remaining -= textLength
        continue
      }

      if (!(node instanceof HTMLElement)) continue
      lastElement = node
      if (node.dataset.editorPlaceholder) continue

      if (node.dataset.templateText) {
        const prefixLength = getTemplatePrefix(node).length
        const textLength = getContentEditableText(node).length
        if (remaining <= prefixLength) {
          setCaretInText(node, 0)
          return
        }
        if (remaining <= prefixLength + textLength) {
          setCaretInText(node, remaining - prefixLength)
          return
        }
        remaining -= prefixLength + textLength
        continue
      }

      const variableLabel = node.dataset.variableLabel
      if (variableLabel) {
        const tokenLength = `{${variableLabel}}`.length
        if (remaining <= tokenLength) {
          setCaretAfter(node)
          return
        }
        remaining -= tokenLength
        continue
      }

      const textLength = serializeTemplate(node.childNodes).length
      if (remaining <= textLength) {
        setTemplateCaret(node.childNodes, remaining)
        return
      }
      remaining -= textLength
    }

    if (lastElement) setCaretAfter(lastElement)
  }

  function setVariableValue(label: string, value: string) {
    if (label === '제목') {
      displayTitle = value
      return
    }

    variableValues = {
      ...variableValues,
      [label]: value,
    }
  }

  function getVariableValue(label: string | undefined) {
    if (!label) return ''
    if (label === '제목') return displayTitle
    return variableValues[label] ?? ''
  }
  function shouldShowReset(label: string | undefined) {
    if (!editable || !label) return false

    if (label === '기존 문서 내용' && isExternalContent && externalContentDirty && onContentResetRequest) return true

    if (!isReservedAutoVariableName(label)) return getVariableValue(label).trim().length > 0

    const defaultValue = reservedDefaults[label]
    if (defaultValue === undefined) return false

    return getVariableValue(label) !== defaultValue
  }

  function getResetTitle(label: string | undefined) {
    if (label === '기존 문서 내용') return '원본 문서 내용으로 되돌리기'
    return label && isReservedAutoVariableName(label) ? '기본값으로 복구' : '값 비우기'
  }

  function isVariableEditable(label: string | undefined) {
    if (!editable || !label) return false
    return true
  }

  function renderReadOnlyPart(part: PromptPart) {
    if (part.type === 'code') return part.templateText ?? part.text
    return part.text
  }

  function getEditableTemplateText(part: PromptPart) {
    return renderReadOnlyPart(part).replace(/^\n{2,}/, (newlines) => newlines.slice(1))
  }

  function getEditableTemplatePrefix(part: PromptPart) {
    const text = renderReadOnlyPart(part)
    return text.startsWith('\n\n') ? '\n' : ''
  }

  function getTemplatePrefix(element: HTMLElement) {
    return element.dataset.templatePrefix ?? ''
  }

  function getSerializedTemplateText(element: HTMLElement) {
    return `${getTemplatePrefix(element)}${getContentEditableText(element)}`
  }
</script>

<div
  {id}
  bind:this={editorEl}
  contenteditable={editable ? 'true' : 'false'}
  role={editable ? 'textbox' : undefined}
  aria-multiline={editable ? 'true' : undefined}
  oninput={editable ? handleEditorInput : undefined}
  onkeydown={editable ? handleEditorKeydown : undefined}
  onpaste={editable ? handleEditorPaste : undefined}
  oncopy={handleEditorCopy}
  oncompositionstart={editable ? handleCompositionStart : undefined}
  oncompositionend={editable ? handleCompositionEnd : undefined}
  onfocusout={editable ? handleEditorFocusOut : undefined}
  class={`z-input flex-1 overflow-y-auto rounded border border-a-slate-300 bg-background p-3 font-mono text-xs outline-none leading-relaxed ${className}`}
>
  {#if liveParts.length === 0}
    <span class="text-a-slate-400 select-none" contenteditable="false" data-editor-placeholder="true">{emptyText}</span>
  {:else}
    <div bind:this={contentEl}>
      {#each liveParts as part, i (i)}
        {#if editable && part.type === 'block' && part.label}
          <div
            class="ai-edit-variable-block rounded border border-dashed border-a-slate-300 bg-a-slate-50/60 p-1"
            contenteditable="false"
            data-variable-label={part.label}
            data-variable-type={part.type}
            data-empty-value={part.text.trim() ? 'false' : 'true'}
            data-value-focused={focusedVariableLabel === part.label ? 'true' : 'false'}
          >
            <div
              class={`flex min-w-0 cursor-text items-center gap-1 ${part.text.trim() || focusedVariableLabel === part.label ? 'text-a-slate-700' : 'text-a-slate-400 select-none'}`}
              data-variable-header="true"
              data-copy-ignore={part.text.trim() ? undefined : 'true'}
              role="button"
              tabindex="0"
              onclick={() => void focusVariableValue(part.label)}
              onkeydown={(event) => handleVariableHeaderKeydown(event, part.label)}
            >
              <span># {part.label}</span>
              {#if shouldShowReset(part.label)}
                <button
                  type="button"
                  class="rounded-full p-0.5 text-a-blue-500 hover:bg-a-blue-50"
                  onmousedown={handleControlPointerDown}
                  onclick={(event) => handleResetClick(event, part)}
                  title={getResetTitle(part.label)}
                >
                  <ZIcon path={mdiRefresh} size="0.7rem" />
                </button>
              {/if}
            </div>
            <div
              class="whitespace-pre-wrap text-a-slate-500 outline-none"
              contenteditable={isVariableEditable(part.label) ? 'true' : 'false'}
              data-variable-value="true"
              role={isVariableEditable(part.label) ? 'textbox' : undefined}
              aria-multiline={isVariableEditable(part.label) ? 'true' : undefined}
              onfocus={() => handleVariableValueFocus(part)}
              onblur={() => handleVariableValueBlur(part)}
            >
              {part.text}
            </div>
          </div>
        {:else if editable && (part.type === 'inline' || part.type === 'block') && part.label}
          <span class="inline-flex max-w-full align-baseline" contenteditable="false" data-variable-label={part.label}>
            <span
              class="relative inline-flex max-w-full items-end gap-1 rounded border border-a-slate-300 bg-a-slate-50 px-1.5 pb-0.5 pt-2"
            >
              <span
                class="absolute -top-1.5 left-1 rounded bg-background px-1 text-[9px] font-semibold leading-none text-a-slate-500 select-none"
                data-copy-ignore="true"
              >
                {part.label}
              </span>
              <span
                class="min-w-6 pr-2 whitespace-pre-wrap text-a-slate-700 outline-none"
                contenteditable={isVariableEditable(part.label) ? 'true' : 'false'}
                data-variable-value="true"
                role={isVariableEditable(part.label) ? 'textbox' : undefined}
              >
                {part.text}
              </span>
              {#if shouldShowReset(part.label)}
                <button
                  type="button"
                  class="shrink-0 rounded-full p-0.5 text-a-blue-500 hover:bg-a-blue-50"
                  onmousedown={handleControlPointerDown}
                  onclick={() => handlePartReset(part)}
                  title={getResetTitle(part.label)}
                >
                  <ZIcon path={mdiRefresh} size="0.7rem" />
                </button>
              {/if}
            </span>
          </span>
        {:else if part.type === 'block' && part.label}
          <div class="my-1">
            <div class="font-semibold text-a-slate-700"># {part.label}</div>
            <div class="whitespace-pre-wrap text-a-slate-500">{renderReadOnlyPart(part)}</div>
          </div>
        {:else if editable}
          <span class="whitespace-pre-wrap" data-template-text="true" data-template-prefix={getEditableTemplatePrefix(part)}>
            {getEditableTemplateText(part)}
          </span>
        {:else}
          <span class="whitespace-pre-wrap">{renderReadOnlyPart(part)}</span>
        {/if}
      {/each}
    </div>
  {/if}
</div>

<style>
  .ai-edit-variable-block[data-empty-value='true'][data-value-focused='false'] [data-variable-value] {
    display: none;
  }
</style>
