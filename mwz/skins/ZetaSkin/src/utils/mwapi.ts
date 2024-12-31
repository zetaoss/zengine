import type MediaWiki from '@wikimedia/types-wikimedia'

import getRLCONF from './rlconf'

async function getMW(): Promise<MediaWiki> {
  let retries = 1
  // @ts-ignore
  while (typeof mw === 'undefined' || typeof mw.Api === 'undefined') {
    await ((t: number) => new Promise((r) => { setTimeout(r, t) }))(retries++)
  }
  // @ts-ignore
  return mw
}

async function getAPI() {
  const mw = await getMW()
  return new mw.Api()
}

export async function getCreated() {
  const pageID = getRLCONF().wgArticleId
  const resp = await (await getAPI()).get({
    action: 'query',
    prop: 'revisions',
    pageids: pageID,
  })
  return resp.query.pages[pageID].revisions[0].timestamp
}

export async function titleExist(title: string) {
  const resp = await (await getAPI()).get({
    action: 'query',
    titles: title,
  })
  return typeof resp.query.pages[-1] !== 'undefined'
}
