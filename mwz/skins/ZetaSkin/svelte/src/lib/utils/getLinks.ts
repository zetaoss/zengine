import type { Link } from '$shared/types/links'
import getShortcut from '$shared/utils/shortcut'

import getRLCONF from './rlconf'

type Filter = string | [key: string, override?: Partial<Link>]
type MaybeFilter = Filter | false | null | undefined

export default function getLinks(...filters: MaybeFilter[]): Link[] {
  const { mm } = getRLCONF()

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
