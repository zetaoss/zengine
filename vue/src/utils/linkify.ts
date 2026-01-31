import Autolinker from 'autolinker'
import DOMPurify from 'isomorphic-dompurify'

import titleExist from './mediawiki'

function linkifyURL(s: string) {
  return Autolinker.link(s, {
    stripPrefix: false,
    sanitizeHtml: false,
    className: 'external',
    urls: { schemeMatches: true, tldMatches: false, ipV4Matches: false },
  })
}

export async function linkifyWikiMatch(s: string, match: string) {
  const title = match.slice(2, -2)
  const exist = await titleExist(title)
  const classList = ['internal', exist ? '' : 'new'].filter(Boolean).join(' ')
  return s
    .split(match)
    .join(`<a href="/wiki/${title}" class="${classList}">${title}</a>`)
}

export async function linkifyWiki(s: string) {
  const matches = Array.from(
    new Set(s.match(/\[\[([^[\]|]*)[^[\]]*\]\]/g) || []),
  )
  for (const match of matches) {
    s = await linkifyWikiMatch(s, match)
  }
  return s
}

export default async function linkify(input: string) {
  const sanitizedInput = DOMPurify.sanitize(input)
  return linkifyWiki(linkifyURL(sanitizedInput))
}
