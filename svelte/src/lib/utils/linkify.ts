import DOMPurify from 'dompurify'

import { linkifyTextOne } from '$shared/utils/linkifyText'
import { titlesExist } from '$shared/utils/mediawiki'
import { extractWikiTitles } from '$shared/utils/wikiLink'

// Links URLs and [[wiki links]] in HTML (forum posts). The HTML is sanitized first; then only its text nodes
// outside links and code are linked, each by linkifyTextOne, which escapes the text and builds the links, so
// markup and attributes are never rewritten. The result is sanitized again.
const skipTags = new Set(['A', 'CODE', 'PRE', 'SCRIPT', 'STYLE', 'TEXTAREA'])
const purifyOptions = { ADD_ATTR: ['target', 'rel'] }

function sanitizedBody(input: string): HTMLElement {
  const html = DOMPurify.sanitize(input, purifyOptions)
  return new DOMParser().parseFromString(`<body>${html}</body>`, 'text/html').body
}

function linkableTexts(root: HTMLElement): Text[] {
  const texts: Text[] = []
  const walker = root.ownerDocument.createTreeWalker(root, NodeFilter.SHOW_TEXT, {
    acceptNode(node) {
      for (let el = node.parentElement; el && el !== root; el = el.parentElement) {
        if (skipTags.has(el.tagName)) return NodeFilter.FILTER_REJECT
      }
      return node.nodeValue?.trim() ? NodeFilter.FILTER_ACCEPT : NodeFilter.FILTER_REJECT
    },
  })
  for (let n = walker.nextNode(); n; n = walker.nextNode()) texts.push(n as Text)
  return texts
}

function linkTitles(input: string): string[] {
  return [...new Set(linkableTexts(sanitizedBody(input)).flatMap((t) => extractWikiTitles(t.nodeValue ?? '')))]
}

export function linkifyHtmlOne(input: string, existsMap: Record<string, boolean>): string {
  const body = sanitizedBody(input)
  for (const text of linkableTexts(body)) {
    const value = text.nodeValue ?? ''
    const linked = linkifyTextOne(value, existsMap)
    if (linked === value) continue
    const template = body.ownerDocument.createElement('template')
    template.innerHTML = linked
    text.replaceWith(template.content)
  }
  return DOMPurify.sanitize(body.innerHTML, purifyOptions)
}

export default async function linkify(inputs: string[]): Promise<string[]> {
  const titles = [...new Set(inputs.flatMap((x) => linkTitles(x || '')))]
  const existsMap = titles.length > 0 ? await titlesExist(titles) : {}
  return inputs.map((x) => linkifyHtmlOne(x || '', existsMap))
}
