import { titlesExist } from '$shared/utils/mediawiki'
import { extractWikiTitles, wikiLinkHtml, wikiLinkRegex } from '$shared/utils/wikiLink'

// Plain-text counterpart of linkify: escapes the text and links only URLs,
// emails and [[wiki links]], so it needs no HTML sanitizer.
const tokenRegex = new RegExp(`(https?://[^\\s<>"'\`]+)|([\\w.+-]+@[\\w-]+(?:\\.[\\w-]+)+)|${wikiLinkRegex.source}`, 'g')
const openers: Record<string, string> = { ')': '(', ']': '[', '}': '{' }

function escapeHtml(s: string) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;')
}

// Ends the URL at its first unbalanced closing bracket, so "(https://a.com)입니다"
// links only https://a.com, and leaves trailing punctuation out.
function trimUrl(url: string) {
  const depth: Record<string, number> = { '(': 0, '[': 0, '{': 0 }
  for (let i = 0; i < url.length; i++) {
    const ch = url[i]
    const opener = openers[ch]
    if (ch in depth) {
      depth[ch]++
    } else if (opener) {
      if (depth[opener] === 0) {
        url = url.slice(0, i)
        break
      }
      depth[opener]--
    }
  }
  return url.replace(/[.,;:!?]+$/, '')
}

export function linkifyTextOne(input: string, existsMap: Record<string, boolean>): string {
  const text = input || ''
  // exec with lastIndex, not matchAll: a trimmed URL hands the rest of its match back
  // to the scan, as in "(https://a.com)[[문서]]".
  const re = new RegExp(tokenRegex.source, 'g')
  let out = ''
  let last = 0
  for (let m = re.exec(text); m; m = re.exec(text)) {
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
      html = target ? wikiLinkHtml(target, escapeHtml((m[4] || m[3] || '').trim()), existsMap[target]) : escapeHtml(raw)
    }
    out += escapeHtml(text.slice(last, m.index)) + html
    last = m.index + raw.length
    re.lastIndex = last
  }
  return out + escapeHtml(text.slice(last))
}

export default async function linkifyText(inputs: string[]): Promise<string[]> {
  const titles = [...new Set(inputs.flatMap((x) => extractWikiTitles(x || '')))]
  const existsMap = titles.length > 0 ? await titlesExist(titles) : {}
  return inputs.map((x) => linkifyTextOne(x || '', existsMap))
}
