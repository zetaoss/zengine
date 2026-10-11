import { titlesExist } from '$shared/utils/mediawiki'
import { getWikiHref } from '$shared/utils/wikiLink'

// Plain-text counterpart of linkify: escapes the text and links only URLs,
// emails and [[wiki links]], so it needs no HTML sanitizer.
const tokenRegex = /(https?:\/\/[^\s<>"'`]+)|([\w.+-]+@[\w-]+(?:\.[\w-]+)+)|\[\[([^\]|]+)(?:\|([^\]]*))?\]\]/g
const wikiTitleRegex = /\[\[([^\]|]+)(?:\|[^\]]*)?\]\]/g
const closers: Record<string, string> = { ')': '(', ']': '[', '}': '{' }

function escapeHtml(s: string) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;')
}

function count(s: string, ch: string) {
  return s.split(ch).length - 1
}

// Leaves trailing punctuation and unbalanced closing brackets out of the URL.
function trimUrl(url: string) {
  for (;;) {
    const last = url.slice(-1)
    const opener = closers[last]
    if (/[.,;:!?]/.test(last) || (opener && count(url, last) > count(url, opener))) {
      url = url.slice(0, -1)
    } else {
      return url
    }
  }
}

export function extractWikiTitles(input: string): string[] {
  const titles = [...(input || '').matchAll(wikiTitleRegex)].map((m) => (m[1] || '').trim()).filter((t) => t.length > 0)
  return [...new Set(titles)]
}

export function linkifyTextOne(input: string, existsMap: Record<string, boolean>): string {
  let out = ''
  let last = 0
  for (const m of (input || '').matchAll(tokenRegex)) {
    const start = m.index ?? 0
    let raw = m[0]
    let html: string
    if (m[1]) {
      raw = trimUrl(m[1])
      const url = escapeHtml(raw)
      html = `<a href="${url}" class="external" target="_blank" rel="nofollow ugc noopener noreferrer">${url}</a>`
    } else if (m[2]) {
      const email = escapeHtml(m[2])
      html = `<a href="mailto:${email}" class="external">${email}</a>`
    } else {
      const target = (m[3] || '').trim()
      if (!target) continue
      const display = (m[4] || m[3] || '').trim()
      const exists = existsMap[target]
      const classList = exists === false ? 'internal new' : 'internal'
      html = `<a href="${escapeHtml(getWikiHref(target, exists))}" class="${classList}" data-sveltekit-reload>${escapeHtml(display)}</a>`
    }
    out += escapeHtml(input.slice(last, start)) + html
    last = start + raw.length
  }
  return out + escapeHtml((input || '').slice(last))
}

export default async function linkifyText(inputs: string[]): Promise<string[]> {
  const titles = [...new Set(inputs.flatMap((x) => extractWikiTitles(x || '')))]
  const existsMap = titles.length > 0 ? await titlesExist(titles) : {}
  return inputs.map((x) => linkifyTextOne(x || '', existsMap))
}
