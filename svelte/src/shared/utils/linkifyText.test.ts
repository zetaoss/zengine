import { describe, expect, it } from 'vitest'

import { linkifyTextOne } from './linkifyText'
import { extractWikiTitles } from './wikiLink'

const ext = (url: string) => `<a href="${url}" class="external" target="_blank" rel="nofollow ugc noopener noreferrer">${url}</a>`

describe('linkifyTextOne', () => {
  it('escapes HTML', () => {
    expect(linkifyTextOne('<정본 소세키 전집> & <b>x</b>', {})).toBe('&lt;정본 소세키 전집&gt; &amp; &lt;b&gt;x&lt;/b&gt;')
    expect(linkifyTextOne('<img src=x onerror=alert(1)>', {})).toBe('&lt;img src=x onerror=alert(1)&gt;')
  })

  it('links URLs', () => {
    expect(linkifyTextOne('see https://wiki.onul.works/ now', {})).toBe(`see ${ext('https://wiki.onul.works/')} now`)
    expect(linkifyTextOne('http://a.com/?x=1&y=2', {})).toBe(ext('http://a.com/?x=1&amp;y=2'))
  })

  it('leaves trailing punctuation and unbalanced brackets out of URLs', () => {
    expect(linkifyTextOne('(https://a.com/x).', {})).toBe(`(${ext('https://a.com/x')}).`)
    expect(linkifyTextOne('https://en.wikipedia.org/wiki/Go_(game),', {})).toBe(`${ext('https://en.wikipedia.org/wiki/Go_(game)')},`)
    expect(linkifyTextOne('링크(https://a.com)입니다', {})).toBe(`링크(${ext('https://a.com')})입니다`)
    expect(linkifyTextOne('[https://a.com/b] 참고', {})).toBe(`[${ext('https://a.com/b')}] 참고`)
    expect(linkifyTextOne('https://a.com/x...', {})).toBe(`${ext('https://a.com/x')}...`)
    expect(linkifyTextOne('https://ko.wikipedia.org/wiki/리눅스 참고', {})).toBe(`${ext('https://ko.wikipedia.org/wiki/리눅스')} 참고`)
  })

  it('keeps scanning after a trimmed URL', () => {
    expect(linkifyTextOne('(https://a.com)[[문서]]', {})).toBe(
      `(${ext('https://a.com')})<a href="/wiki/%EB%AC%B8%EC%84%9C" class="internal" data-sveltekit-reload>문서</a>`,
    )
    expect(linkifyTextOne('[https://a.com][https://b.com]', {})).toBe(`[${ext('https://a.com')}][${ext('https://b.com')}]`)
  })

  it('does not break out of attributes', () => {
    expect(linkifyTextOne('https://a.com/"onmouseover="x', {})).toBe(`${ext('https://a.com/')}&quot;onmouseover=&quot;x`)
    expect(linkifyTextOne('javascript:alert(1)', {})).toBe('javascript:alert(1)')
  })

  it('does not link emails', () => {
    expect(linkifyTextOne('mail a.b@example.com', {})).toBe('mail a.b@example.com')
  })

  it('links wiki titles', () => {
    expect(linkifyTextOne('[[리눅스]] [[없는 문서|표시]] [[A|]]', { 리눅스: true, '없는 문서': false })).toBe(
      '<a href="/wiki/%EB%A6%AC%EB%88%85%EC%8A%A4" class="internal" data-sveltekit-reload>리눅스</a> ' +
        '<a href="/w/index.php?title=%EC%97%86%EB%8A%94_%EB%AC%B8%EC%84%9C&amp;action=edit&amp;redlink=1" class="internal new" data-sveltekit-reload>표시</a> ' +
        '<a href="/wiki/A" class="internal" data-sveltekit-reload>A</a>',
    )
    expect(linkifyTextOne('[[<b>|<i>]]', {})).toBe('<a href="/wiki/%3Cb%3E" class="internal" data-sveltekit-reload>&lt;i&gt;</a>')
  })
})

describe('extractWikiTitles', () => {
  it('dedupes and trims', () => {
    expect(extractWikiTitles('[[ A ]] [[A|x]] [[B]]')).toEqual(['A', 'B'])
  })
})
