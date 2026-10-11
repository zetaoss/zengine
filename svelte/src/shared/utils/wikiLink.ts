function encodeWikiPathTitle(title: string): string {
  return title
    .replace(/ /g, '_')
    .split('/')
    .map((segment) => encodeURIComponent(segment))
    .join('/')
}

function encodeWikiQueryTitle(title: string): string {
  return encodeURIComponent(title.replace(/ /g, '_'))
}

export function getWikiViewHref(title: string): string {
  return `/wiki/${encodeWikiPathTitle(title)}`
}

export function getWikiHref(title: string, exists?: boolean): string {
  return exists === false ? `/w/index.php?title=${encodeWikiQueryTitle(title)}&action=edit&redlink=1` : getWikiViewHref(title)
}

export function getWikiEditHref(title: string): string {
  return `/w/index.php?title=${encodeWikiQueryTitle(title)}&action=edit`
}

export function getWikiDiffHref(title: string, revid?: number): string {
  if (revid && revid > 0) {
    return `/w/index.php?title=${encodeWikiQueryTitle(title)}&diff=${revid}&oldid=prev`
  }
  return `/w/index.php?title=${encodeWikiQueryTitle(title)}&diff=cur&oldid=prev`
}

export const wikiLinkRegex = /\[\[([^\]|]+)(?:\|([^\]]*))?\]\]/g

export function extractWikiTitles(input: string): string[] {
  const titles = [...(input || '').matchAll(wikiLinkRegex)].map((m) => (m[1] || '').trim()).filter((t) => t.length > 0)
  return [...new Set(titles)]
}

// displayHtml must already be safe HTML.
export function wikiLinkHtml(target: string, displayHtml: string, exists?: boolean): string {
  const href = getWikiHref(target, exists).replace(/&/g, '&amp;')
  const classList = exists === false ? 'internal new' : 'internal'
  return `<a href="${href}" class="${classList}" data-sveltekit-reload>${displayHtml}</a>`
}
