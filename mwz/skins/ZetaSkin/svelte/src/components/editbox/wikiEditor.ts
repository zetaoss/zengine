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
  textarea.setSelectionRange(0, textarea.value.length)
   
  const inserted = document.execCommand('insertText', false, content)
  if (!inserted) {
    textarea.value = content
    textarea.dispatchEvent(new Event('input', { bubbles: true }))
  }
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
