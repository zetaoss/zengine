import type { Avatar } from '@common/components/avatar/avatar'

import type { Binder } from '@/components/binder/types'

interface RLCONF {
  wgArticleId: number,
  wgCategories: string[],
  wgUserId: number,
  wgUserGroups: string[],
  avatar: Avatar,
  binders: Binder[],
  contributors: Avatar[],
  lastmod: string,
}

export default function getRLCONF(): RLCONF {
  // @ts-expect-error: external variable RLCONF
  return RLCONF
}
