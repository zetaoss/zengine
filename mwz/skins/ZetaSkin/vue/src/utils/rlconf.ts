import type UserAvatar from '@common/types/userAvatar'

import type { Binder } from '@/components/binder/types'

interface RLCONF {
  wgArticleId: number,
  wgCategories: string[],
  wgUserId: number,
  wgUserGroups: string[],
  avatar: UserAvatar,
  binders: Binder[],
  contributors: UserAvatar[],
  lastmod: string,
}

export default function getRLCONF(): RLCONF {
  // @ts-ignore
  return RLCONF
}
