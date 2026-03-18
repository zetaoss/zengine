function encodeWikiTitle(title: string): string {
  return encodeURIComponent(title.replace(/ /g, '_'))
}

function getWikiEditHref(title: string): string {
  return `/w/index.php?title=${encodeWikiTitle(title)}&action=edit&redlink=1`
}

export function getWikiViewHref(title: string): string {
  return `/wiki/${encodeWikiTitle(title)}`
}

export function getWikiHref(title: string, exists?: boolean): string {
  return exists === false ? getWikiEditHref(title) : getWikiViewHref(title)
}
