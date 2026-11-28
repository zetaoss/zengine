import { describe, expect, it } from 'vitest'

import { getWikiHref, getWikiViewHref } from './wikiLink'

describe('getWikiViewHref', () => {
  it('encodes characters a URL path would read as its end or a query', () => {
    expect(getWikiViewHref('CAROLE KING의 컨셉트 앨범?!')).toBe(
      '/wiki/CAROLE_KING%EC%9D%98_%EC%BB%A8%EC%85%89%ED%8A%B8_%EC%95%A8%EB%B2%94%3F!',
    )
    expect(getWikiViewHref('C#')).toBe('/wiki/C%23')
    expect(getWikiViewHref('100%')).toBe('/wiki/100%25')
    expect(getWikiViewHref('A&B')).toBe('/wiki/A%26B')
  })

  it('keeps subpage slashes and turns spaces into underscores', () => {
    expect(getWikiViewHref('틀:새문서틀 인물/설명')).toBe('/wiki/%ED%8B%80%3A%EC%83%88%EB%AC%B8%EC%84%9C%ED%8B%80_%EC%9D%B8%EB%AC%BC/%EC%84%A4%EB%AA%85')
  })
})

describe('getWikiHref', () => {
  it('links a missing page to its edit form with the title encoded', () => {
    expect(getWikiHref('A?B', false)).toBe('/w/index.php?title=A%3FB&action=edit&redlink=1')
    expect(getWikiHref('A?B', true)).toBe('/wiki/A%3FB')
  })
})
