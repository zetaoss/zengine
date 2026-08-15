export function findWikiEditor(): HTMLTextAreaElement | null {
  return document.querySelector<HTMLTextAreaElement>('#wpTextbox1')
}

export function getWikiEditorContent(): string {
  return findWikiEditor()?.value ?? ''
}

export function replaceWikiEditorContent(content: string): boolean {
  const textarea = findWikiEditor()
  if (!textarea) return false

  textarea.focus()
  textarea.select()
  try {
    if (document.execCommand('insertText', false, content)) return true
  } catch {
    // Fall back to setting the value when the browser does not support insertText.
  }

  textarea.value = content
  textarea.dispatchEvent(new Event('input', { bubbles: true }))
  return true
}

export function subscribeWikiEditorContent(onChange: (content: string) => void): () => void {
  const textarea = findWikiEditor()
  if (!textarea) return () => undefined

  const handleInput = () => onChange(textarea.value)
  handleInput()
  textarea.addEventListener('input', handleInput)
  return () => textarea.removeEventListener('input', handleInput)
}
