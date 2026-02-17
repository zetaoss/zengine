import type { Binder } from '$lib/types/binder'
import type { Contributor } from '$lib/types/contributor'
import type { LinkMapMap } from '$shared/types/links'
import type { DataToc } from '$lib/types/toc'

type RLConfig = {
  binders: Binder[]
  contributors: Contributor[]
  datatoc: DataToc
  mm: LinkMapMap
  revtime: string
  wgAction: string
  wgArticleId: number
  wgUserGroups: string[]
  wgUserId: number
  wgUserName: string
}

export default function getRLCONF(): RLConfig {
  return (globalThis as unknown as { RLCONF: RLConfig }).RLCONF
}
