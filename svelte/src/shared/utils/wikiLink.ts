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
