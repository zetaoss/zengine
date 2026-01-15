// @/src/utils/rlconf.ts
import type { User } from '@common/types/user'

import type { Binder } from '@/components/binder/types'

interface RLCONF {
  wgArticleId: number,
  wgCategories: string[],
  wgUserId: number,
  wgUserName: string,
  wgUserGroups: string[],
  binders: Binder[],
  lastmod: string,
  contributors: User[],
}

export default function getRLCONF(): RLCONF {
  // @ts-expect-error: external variable RLCONF
  return RLCONF
}
