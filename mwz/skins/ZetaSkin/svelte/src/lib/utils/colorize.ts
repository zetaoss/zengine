const PATTERN = /#[0-9a-fA-F]{3,8}\b|rgba?\([^)\n]+\)|hsla?\([^)\n]+\)|\b[a-zA-Z-]+\b/g
const KEYWORDS = ['inherit', 'initial', 'unset', 'revert', 'revert-layer', 'transparent', 'currentcolor']

export const COLORIZE_CLASS = 'colorize'

function isColorToken(token: string): boolean {
  if (typeof CSS === 'undefined' || typeof CSS.supports !== 'function') return false
  const normalized = token.trim()
  if (!normalized) return false
  if (KEYWORDS.includes(normalized.toLowerCase())) return false
  return CSS.supports('color', normalized)
}

function getStyleContentRanges(fullText: string): Array<{ start: number; end: number }> {
  const ranges: Array<{ start: number; end: number }> = []
  const openRe = /<style\b[^>]*>/gi
  const closeRe = /<\/style\s*>/gi

  let openMatch = openRe.exec(fullText)
  while (openMatch) {
    const contentStart = openRe.lastIndex
    closeRe.lastIndex = contentStart
    const closeMatch = closeRe.exec(fullText)
    if (!closeMatch) break

    ranges.push({ start: contentStart, end: closeMatch.index })
    openRe.lastIndex = closeRe.lastIndex
    openMatch = openRe.exec(fullText)
  }

  return ranges
}

function render(sourceHtml: string, shouldApply: (token: string, index: number, fullText: string) => boolean): string {
  if (typeof document === 'undefined') return sourceHtml
  if (!sourceHtml) return sourceHtml

  const root = document.createElement('div')
  root.innerHTML = sourceHtml

  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT)
  const textNodes: Array<{ node: Text; start: number; end: number }> = []
  let cursor = 0

  while (walker.nextNode()) {
    const node = walker.currentNode as Text
    const start = cursor
    const end = start + node.data.length
    textNodes.push({ node, start, end })
    cursor = end
  }

  const fullText = textNodes.map((t) => t.node.data).join('')
  const matches: Array<{ index: number; token: string }> = []
  PATTERN.lastIndex = 0
  let match = PATTERN.exec(fullText)

  while (match) {
    matches.push({ index: match.index, token: match[0] })
    match = PATTERN.exec(fullText)
  }

  for (let i = matches.length - 1; i >= 0; i -= 1) {
    const { index, token } = matches[i]
    if (!isColorToken(token)) continue
    if (!shouldApply(token, index, fullText)) continue

    const targetInfo = textNodes.find((t) => index >= t.start && index < t.end)
    if (!targetInfo) continue

    const offsetInNode = index - targetInfo.start
    let targetNode = targetInfo.node
    if (offsetInNode > 0) {
      targetNode = targetNode.splitText(offsetInNode)
    }

    const swatch = document.createElement('span')
    swatch.className = COLORIZE_CLASS
    swatch.style.backgroundColor = token
    swatch.title = token
    targetNode.parentNode?.insertBefore(swatch, targetNode)
  }

  return root.innerHTML
}

export function colorizeCss(sourceHtml: string): string {
  return render(sourceHtml, () => true)
}

export function colorizeHtml(sourceHtml: string): string {
  let ranges: Array<{ start: number; end: number }> | null = null
  return render(sourceHtml, (_token, index, fullText) => {
    if (!ranges) ranges = getStyleContentRanges(fullText)
    return ranges.some((range) => index >= range.start && index < range.end)
  })
}
