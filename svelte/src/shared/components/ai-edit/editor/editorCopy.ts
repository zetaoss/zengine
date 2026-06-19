function normalizeCopiedText(text: string) {
  return text.replace(/\u00a0/g, ' ').replace(/\n{3,}/g, '\n\n')
}

export function copyEditorSelection(contentEl: HTMLElement, event: ClipboardEvent) {
  const selection = window.getSelection()
  if (!selection || selection.rangeCount === 0 || selection.isCollapsed) return

  const range = selection.getRangeAt(0)
  if (!contentEl.contains(range.commonAncestorContainer)) return

  const fragment = range.cloneContents()
  for (const variableEl of Array.from(fragment.querySelectorAll<HTMLElement>('[data-variable-label]'))) {
    const label = variableEl.dataset.variableLabel
    const type = variableEl.dataset.variableType
    const value = variableEl.querySelector<HTMLElement>('[data-variable-value]')?.innerText ?? ''
    const copiedValue = type === 'block' && label && value.trim() ? `# ${label}\n${value}` : value
    variableEl.replaceWith(document.createTextNode(copiedValue))
  }
  for (const ignoredEl of Array.from(fragment.querySelectorAll<HTMLElement>('[data-copy-ignore]'))) {
    ignoredEl.remove()
  }

  const container = document.createElement('div')
  container.appendChild(fragment)
  event.preventDefault()
  event.clipboardData?.setData('text/plain', normalizeCopiedText(container.innerText || container.textContent || ''))
}
