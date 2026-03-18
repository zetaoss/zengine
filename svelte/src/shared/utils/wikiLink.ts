function normalizeWikiTitle(title: string): string {
  return title.replace(/ /g, '_')
}

function encodeWikiPathTitle(title: string): string {
  return normalizeWikiTitle(title)
    .split('/')
    .map((segment) => encodeURIComponent(segment))
    .join('/')
}

function encodeWikiQueryTitle(title: string): string {
  return encodeURIComponent(title.replace(/ /g, '_'))
}

function getWikiEditHref(title: string): string {
  return `/w/index.php?title=${encodeWikiQueryTitle(title)}&action=edit&redlink=1`
}

export function getWikiViewHref(title: string): string {
  return `/wiki/${encodeWikiPathTitle(title)}`
}

export function getWikiHref(title: string, exists?: boolean): string {
  return exists === false ? getWikiEditHref(title) : getWikiViewHref(title)
}
