export type MyMenuItem = {
  accesskey?: string
  active?: boolean
  class?: string | false
  'data-mw'?: string
  exists?: boolean
  href: string
  icon?: string
  id?: string
  'link-class'?: string[]
  'single-id'?: string
  text: string
  title?: string
}

export type MyMenu = Record<string, MyMenuItem>

type RLConf = {
  binders: unknown[]
  contributors: Array<{ id: number; name: string }>
  dataToc: unknown
  lastmod: string
  myMenu: MyMenu
  wgArticleId: number
  wgUserGroups: string[]
  wgUserId: number
  wgUserName: string
}

export default function getRLCONF(): RLConf {
  // In this app, RLCONF is always present and already normalized by the server.
  const c = (globalThis as unknown as { RLCONF: Record<string, unknown> }).RLCONF

  const binders = (c.binders ?? []) as unknown[]
  const contributors = (c.contributors ?? []) as Array<{ id: number; name: string }>
  const dataToc = c.dataToc ?? []
  const lastmod = (c.lastmod ?? '') as string
  const myMenu = (c.myMenu ?? {}) as MyMenu
  const wgArticleId = (c.wgArticleId ?? 0) as number
  const wgUserGroups = (c.wgUserGroups ?? []) as string[]
  const wgUserId = (c.wgUserId ?? 0) as number
  const wgUserName = (c.wgUserName ?? '') as string

  return { binders, contributors, dataToc, lastmod, myMenu, wgArticleId, wgUserGroups, wgUserId, wgUserName }
}
