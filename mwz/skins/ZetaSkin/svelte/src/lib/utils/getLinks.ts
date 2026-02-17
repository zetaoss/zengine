import type { Link } from '$lib/types/links'

import getRLCONF from './rlconf'

type Filter = string | [key: string, override?: Partial<Link>]
type MaybeFilter = Filter | false | null | undefined

function getModifier(): string {
  if (typeof navigator === 'undefined') return 'Alt'

  const ua = navigator.userAgent
  const isMac = /Macintosh|Mac OS X|iPhone|iPad|iPod/i.test(ua)
  const isFirefox = /Firefox\//i.test(ua)

  if (isMac) return 'Ctrl+Option'
  if (isFirefox) return 'Alt+Shift'
  return 'Alt'
}

function getShortcut(accesskey?: string): string | undefined {
  const key = accesskey?.trim()
  if (!key) return undefined
  const display = key.length === 1 ? key.toUpperCase() : key
  return `${getModifier()}+${display}`
}

export default function getLinks(...filters: MaybeFilter[]): Link[] {
  const { mm } = getRLCONF()
  console.log('mm', mm)

  return filters.flatMap((filter) => {
    if (!filter) return []

    const [path, override] = typeof filter === 'string' ? [filter, undefined] : filter
    const [group, key] = path.split('.', 2)
    const item = mm[group]?.[key]
    if (!item) return []

    const link = { ...item, ...(override ?? {}) }
    link.title = link.title || link.text

    const shortcut = getShortcut(link.accesskey)
    if (shortcut && link.title) {
      link.title = `${link.title} (${shortcut})`
    }

    return [link]
  })
}
