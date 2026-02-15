type RLConf = {
  binders: unknown[]
  contributors: Array<{ id: number; name: string }>
  dataToc: unknown
  lastmod: string
  wgArticleId: number
  wgUserGroups: string[]
  wgUserId: number
  wgUserName: string
}

export default function getRLCONF(): RLConf {
  const globalConf = (globalThis as { RLCONF?: Record<string, unknown> }).RLCONF
  const mwObj = (globalThis as { mw?: unknown }).mw as
    | {
        config?: {
          get?: (key: string) => unknown
        }
      }
    | undefined

  const read = (key: string): unknown => {
    if (globalConf && key in globalConf) {
      return globalConf[key]
    }

    const config = mwObj?.config
    if (!config || typeof config.get !== 'function') return undefined
    try {
      // MediaWiki's get() depends on `this` being config object.
      return config.get.call(config, key)
    } catch {
      return undefined
    }
  }

  const wgArticleId = Number(read('wgArticleId') ?? 0)
  const wgUserId = Number(read('wgUserId') ?? 0)
  const wgUserName = String(read('wgUserName') ?? '')
  const wgUserGroupsRaw = read('wgUserGroups')
  const wgUserGroups = Array.isArray(wgUserGroupsRaw) ? wgUserGroupsRaw.map((v) => String(v)) : []
  const bindersRaw = read('binders')
  const binders = Array.isArray(bindersRaw) ? bindersRaw : []
  const lastmod = String(read('lastmod') ?? '')
  const contributorsRaw = read('contributors')
  const contributors = Array.isArray(contributorsRaw) ? (contributorsRaw as Array<{ id: number; name: string }>) : []
  const dataToc = read('dataToc') ?? []

  return { wgArticleId, wgUserId, wgUserName, wgUserGroups, binders, lastmod, contributors, dataToc }
}
