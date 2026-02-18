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
  const wikiLinkRegex = /\[\[([^\]|]+)(?:\|([^\]]*))?\]\]/g
  const matches = [...s.matchAll(wikiLinkRegex)]
  if (matches.length === 0) return s

  const titles = [...new Set(matches.map((match) => (match[1] || '').trim()))]
  const existsMap = await titlesExist(titles)

  return s.replace(wikiLinkRegex, (_match, targetRaw: string, displayRaw: string | undefined) => {
    const target = (targetRaw || '').trim()
    const display = (displayRaw || targetRaw).trim()
    const classList = existsMap[target] === true ? 'internal' : 'internal new'
    const href = `/wiki/${encodeURIComponent(target.replace(/ /g, '_'))}`
    return `<a href="${href}" class="${classList}" data-sveltekit-reload>${display}</a>`
  })
}

export default async function linkify(input: string) {
  const linked = await linkifyWiki(linkifyURL(input))
  return DOMPurify.sanitize(linked, { ADD_ATTR: ['target', 'rel'] })
}
