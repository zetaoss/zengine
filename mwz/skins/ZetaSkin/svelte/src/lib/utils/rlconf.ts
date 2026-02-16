import type { Binder } from '$lib/types/binder'
import type { Contributor } from '$lib/types/contributor'
import type { MyMenu, PageMenu } from '$lib/types/menu'
import type { DataToc } from '$lib/types/toc'

type RLConfig = {
  binders: Binder[]
  contributors: Contributor[]
  dataToc: DataToc
  lastmod: string
  myMenu: MyMenu
  pageMenu: PageMenu
  wgArticleId: number
  wgUserGroups: string[]
  wgUserId: number
  wgUserName: string
}

export default function getRLCONF(): RLConfig {
  // In this app, RLCONF is always present and already normalized by the server.
  const c = (globalThis as unknown as { RLCONF: Partial<RLConfig> }).RLCONF

  return {
    binders: c.binders ?? [],
    contributors: c.contributors ?? [],
    dataToc: c.dataToc ?? { 'array-sections': [] },
    lastmod: c.lastmod ?? '',
    myMenu: c.myMenu ?? {},
    pageMenu: c.pageMenu ?? [],
    wgArticleId: c.wgArticleId ?? 0,
    wgUserGroups: c.wgUserGroups ?? [],
    wgUserId: c.wgUserId ?? 0,
    wgUserName: c.wgUserName ?? '',
  }
}
