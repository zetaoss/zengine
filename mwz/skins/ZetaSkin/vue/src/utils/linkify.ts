import Autolinker from 'autolinker'
import DOMPurify from "isomorphic-dompurify";

import { titleExist } from './mwapi'

function linkifyURL(s: string) {
  return Autolinker.link(s, {
    stripPrefix: false,
    sanitizeHtml: false,
    className: 'external',
    urls: { schemeMatches: true, tldMatches: false, ipV4Matches: false },
  })
}

async function linkifyWikiMatch(s: string, match: string) {
  const title = match.slice(2, -2)
  const exist: boolean = await titleExist(title)
  const newClass = exist ? 'new' : ''
  return s.split(match).join(`<a href="/wiki/${title}" class="internal ${newClass}">${title}</a>`)
}

async function linkifyWiki(s: string) {
  const temp = s.match(/\[\[([^[\]|]*)[^[\]]*\]\]/g)
  if (!temp) return s
  const matches = [...new Set(temp)]
  for (const match of matches) {
    s = await linkifyWikiMatch(s, match)
  }
  return s
}

export default async function linkify(input: string) {
  return linkifyWiki(linkifyURL(DOMPurify.sanitize(input)))
}
