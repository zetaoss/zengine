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
