import { describe, expect, it } from 'vitest'

import { linkifyHtmlOne } from './linkify'

const ext = (url: string) => `<a href="${url}" class="external" target="_blank" rel="nofollow ugc noopener noreferrer">${url}</a>`
const wiki = (path: string, text: string) => `<a href="/wiki/${path}" class="internal" data-sveltekit-reload="">${text}</a>`

describe('linkifyHtmlOne', () => {
  it('links URLs in text and keeps the markup around them', () => {
    expect(linkifyHtmlOne('<p>see <b>https://a.com/x</b>, now</p>', {})).toBe(`<p>see <b>${ext('https://a.com/x')}</b>, now</p>`)
    expect(linkifyHtmlOne('<p>(https://a.com/?x=1&amp;y=2)</p>', {})).toBe(`<p>(${ext('https://a.com/?x=1&amp;y=2')})</p>`)
  })

  it('leaves links, code and attributes alone', () => {
    expect(linkifyHtmlOne('<a href="https://b.com">https://b.com</a>', {})).toBe('<a href="https://b.com">https://b.com</a>')
    expect(linkifyHtmlOne('<pre><code>curl https://a.com [[문서]]</code></pre>', {})).toBe('<pre><code>curl https://a.com [[문서]]</code></pre>')
    expect(linkifyHtmlOne('<p><code>https://a.com</code> https://b.com</p>', {})).toBe(`<p><code>https://a.com</code> ${ext('https://b.com')}</p>`)
    expect(linkifyHtmlOne('<img src="https://a.com/i.png" alt="[[문서]] https://b.com">', {})).toBe(
      '<img src="https://a.com/i.png" alt="[[문서]] https://b.com">',
    )
  })

  it('links wiki titles in text', () => {
    expect(linkifyHtmlOne('<p>[[리눅스]] 참고</p>', { 리눅스: true })).toBe(`<p>${wiki('%EB%A6%AC%EB%88%85%EC%8A%A4', '리눅스')} 참고</p>`)
  })

  it('escapes text that reads as markup once linked', () => {
    expect(linkifyHtmlOne('<p>&lt;img src=x onerror=alert(1)&gt; https://a.com</p>', {})).toBe(
      `<p>&lt;img src=x onerror=alert(1)&gt; ${ext('https://a.com')}</p>`,
    )
    expect(linkifyHtmlOne('<p>https://a.com/"onmouseover="x</p>', {})).toBe(`<p>${ext('https://a.com/')}"onmouseover="x</p>`)
  })

  it('removes what the sanitizer removes', () => {
    expect(linkifyHtmlOne('<p onclick="x()">a</p><script>alert(1)</script>', {})).toBe('<p>a</p>')
    expect(linkifyHtmlOne('<a href="javascript:alert(1)">x</a> javascript:alert(1)', {})).toBe('<a>x</a> javascript:alert(1)')
  })
})
