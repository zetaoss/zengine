import Autolinker from 'autolinker'
import DOMPurify from 'isomorphic-dompurify'

import { titlesExist } from './mediawiki'

function linkifyURL(s: string) {
  return Autolinker.link(s, {
    stripPrefix: false,
    sanitizeHtml: false,
    className: 'external',
    urls: { schemeMatches: true, tldMatches: false, ipV4Matches: false },
  })
}

export function linkifyWikiMatch(s: string, match: string, existsMap: Record<string, boolean>) {
  const title = match.slice(2, -2)
  const exist = existsMap[title] === true
  const classList = ['internal', exist ? '' : 'new'].filter(Boolean).join(' ')
  return s.split(match).join(`<a href="/wiki/${title}" class="${classList}" data-sveltekit-reload>${title}</a>`)
}

export async function linkifyWiki(s: string) {
  const matches = Array.from(new Set(s.match(/\[\[([^[\]|]*)[^[\]]*\]\]/g) || []))
  if (matches.length === 0) return s
  const titles = matches.map((match) => match.slice(2, -2))
  const existsMap = await titlesExist(titles)
  for (const match of matches) {
    s = linkifyWikiMatch(s, match, existsMap)
  }
  return s
}

export default async function linkify(input: string) {
  const sanitizedInput = DOMPurify.sanitize(input)
  return linkifyWiki(linkifyURL(sanitizedInput))
}
