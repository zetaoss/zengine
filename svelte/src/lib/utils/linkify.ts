import Autolinker from 'autolinker'
import DOMPurify from 'isomorphic-dompurify'

import { titlesExist } from '$shared/utils/mediawiki'
import { extractWikiTitles, wikiLinkHtml, wikiLinkRegex } from '$shared/utils/wikiLink'

function linkifyURL(s: string) {
  return Autolinker.link(s, {
    stripPrefix: false,
    sanitizeHtml: false,
    className: 'external',
    urls: { schemeMatches: true, tldMatches: false, ipV4Matches: false },
  })
}

function linkifyWiki(s: string, existsMap: Record<string, boolean>) {
  return s.replace(wikiLinkRegex, (_match, targetRaw: string, displayRaw: string | undefined) => {
    const target = (targetRaw || '').trim()
    return wikiLinkHtml(target, (displayRaw || targetRaw).trim(), existsMap[target])
  })
}

async function linkifyOne(input: string, existsMap: Record<string, boolean>) {
  const linked = linkifyWiki(linkifyURL(input), existsMap)
  return DOMPurify.sanitize(linked, { ADD_ATTR: ['target', 'rel'] })
}

export default async function linkify(inputs: string[]): Promise<string[]> {
  const titles = [...new Set(inputs.flatMap((x) => extractWikiTitles(x || '')))]
  const existsMap = titles.length > 0 ? await titlesExist(titles) : {}
  return Promise.all(inputs.map((x) => linkifyOne(x || '', existsMap)))
}
