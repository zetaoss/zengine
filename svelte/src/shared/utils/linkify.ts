import Autolinker from 'autolinker'
import DOMPurify from 'isomorphic-dompurify'

import { titlesExist } from '$shared/utils/mediawiki'

const wikiLinkRegex = /\[\[([^\]|]+)(?:\|([^\]]*))?\]\]/g

function linkifyURL(s: string) {
  return Autolinker.link(s, {
    stripPrefix: false,
    sanitizeHtml: false,
    className: 'external',
    urls: { schemeMatches: true, tldMatches: false, ipV4Matches: false },
  })
}

function extractWikiTitles(input: string): string[] {
  const matches = [...(input || '').matchAll(wikiLinkRegex)]
  if (matches.length === 0) return []
  return [...new Set(matches.map((match) => (match[1] || '').trim()).filter((title) => title.length > 0))]
}

async function linkifyWiki(s: string, existsMap: Record<string, boolean>) {
  const titles = extractWikiTitles(s)
  if (titles.length === 0) return s

  return s.replace(wikiLinkRegex, (_match, targetRaw: string, displayRaw: string | undefined) => {
    const target = (targetRaw || '').trim()
    const display = (displayRaw || targetRaw).trim()
    const classList = existsMap[target] === true ? 'internal' : 'internal new'
    const href = `/wiki/${encodeURIComponent(target.replace(/ /g, '_'))}`
    return `<a href="${href}" class="${classList}" data-sveltekit-reload>${display}</a>`
  })
}

async function linkifyOne(input: string, existsMap: Record<string, boolean>) {
  const linked = await linkifyWiki(linkifyURL(input), existsMap)
  return DOMPurify.sanitize(linked, { ADD_ATTR: ['target', 'rel'] })
}

export default async function linkify(inputs: string[]): Promise<string[]> {
  const titles = [...new Set(inputs.flatMap((x) => extractWikiTitles(x || '')))]
  const existsMap = titles.length > 0 ? await titlesExist(titles) : {}
  return Promise.all(inputs.map((x) => linkifyOne(x || '', existsMap)))
}
