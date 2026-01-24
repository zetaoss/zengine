// @/src/utils/rlconf.ts
import type { User } from '@common/types/user'

import type { Binder } from '@/types/binder'
import type { Section } from '@/types/toc'

interface RLCONF {
  wgArticleId: number,
  wgCategories: string[],
  wgUserId: number,
  wgUserName: string,
  wgUserGroups: string[],
  binders: Binder[],
  lastmod: string,
  contributors: User[],
  dataToc: Section,
}

export default function getRLCONF(): RLCONF {
  // @ts-expect-error: external variable RLCONF
  return RLCONF
}
