export function getNodeElement(node: Node) {
  if (node instanceof HTMLElement) return node
  return node.parentNode instanceof HTMLElement ? node.parentNode : undefined
}

export function getContentEditableText(element: HTMLElement) {
  return element.innerText.replace(/\u00a0/g, ' ')
}

export function getTextOffset(root: HTMLElement, focusNode: Node, focusOffset: number) {
  const range = document.createRange()
  range.selectNodeContents(root)
  try {
    range.setEnd(focusNode, focusOffset)
  } catch {
    return getContentEditableText(root).length
  }
  return range.toString().length
}

export function setCaretInText(root: HTMLElement, targetOffset: number) {
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT)
  let remaining = targetOffset
  let textNode = walker.nextNode()

  while (textNode) {
    const textLength = textNode.textContent?.length ?? 0
    if (remaining <= textLength) {
      setCaret(textNode, remaining)
      return
    }
    remaining -= textLength
    textNode = walker.nextNode()
  }

  const fallbackNode = document.createTextNode('')
  root.appendChild(fallbackNode)
  setCaret(fallbackNode, 0)
}

export function setCaret(node: Node, offset: number) {
  const selection = window.getSelection()
  if (!selection) return

  const range = document.createRange()
  range.setStart(node, offset)
  range.collapse(true)
  selection.removeAllRanges()
  selection.addRange(range)
}

export function setCaretAfter(node: Node) {
  const selection = window.getSelection()
  if (!selection) return

  const range = document.createRange()
  range.setStartAfter(node)
  range.collapse(true)
  selection.removeAllRanges()
  selection.addRange(range)
}
