import Autolinker from 'autolinker'
import DOMPurify from 'isomorphic-dompurify'

import { titlesExist } from '$shared/utils/mediawiki'

function linkifyURL(s: string) {
  return Autolinker.link(s, {
    stripPrefix: false,
    sanitizeHtml: false,
    className: 'external',
    urls: { schemeMatches: true, tldMatches: false, ipV4Matches: false },
  })
}

async function linkifyWiki(s: string) {
  const matches = [...new Set(s.match(/\[\[([^[\]|]*)[^[\]]*\]\]/g) ?? [])]
  if (matches.length === 0) return s
  const titles = matches.map((match) => (match.slice(2, -2).split('|', 2)[0] || '').trim())
  const existsMap = await titlesExist(titles)

  for (const match of matches) {
    const [targetRaw, displayRaw] = match.slice(2, -2).split('|', 2)
    const target = (targetRaw || '').trim()
    const display = (displayRaw || target).trim()
    const classList = existsMap[target] === true ? 'internal' : 'internal new'
    const href = `/wiki/${encodeURIComponent(target.replace(/ /g, '_'))}`
    s = s.split(match).join(`<a href="${href}" class="${classList}" data-sveltekit-reload>${display}</a>`)
  }

  return s
}

export default async function linkify(input: string) {
  const linked = await linkifyWiki(linkifyURL(input))
  return DOMPurify.sanitize(linked)
}
