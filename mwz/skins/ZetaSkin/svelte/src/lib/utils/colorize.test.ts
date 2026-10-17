import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { COLORIZE_CLASS, colorizeCss, colorizeHtml } from './colorize'

describe('colorize', () => {
  beforeEach(() => {
    vi.stubGlobal('CSS', {
      supports: (prop: string, value: string) => {
        if (prop !== 'color') return false
        const v = value.trim().toLowerCase()
        return (
          v === 'skyblue' ||
          v === 'red' ||
          v === '#ace' ||
          v === '#ff0005' ||
          v === '#f005' ||
          v === 'rgb(255 255 255)' ||
          v === 'rgb(255, 255, 255)'
        )
      },
    })
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('adds one swatch for split rgb token in css-highlight html', () => {
    const source = '<span>div</span>{<span>color</span>:<span>rgb</span>(<span>255</span> <span>255</span> <span>255</span>);}'

    const before = (() => {
      const el = document.createElement('div')
      el.innerHTML = source
      return el.textContent
    })()

    const rendered = colorizeCss(source)
    const out = document.createElement('div')
    out.innerHTML = rendered

    expect(out.textContent).toBe(before)
    expect(out.querySelectorAll(`.${COLORIZE_CLASS}`)).toHaveLength(1)
    expect(out.innerHTML).toContain('title="rgb(255 255 255)"')
  })

  it('adds preview only for style content when html mode is used', () => {
    const source = '<span>&lt;style&gt;</span><span>.x { color: skyblue; }</span><span>&lt;/style&gt;</span><span> skyblue </span>'
    const rendered = colorizeHtml(source)
    const out = document.createElement('div')
    out.innerHTML = rendered

    expect(out.querySelectorAll(`.${COLORIZE_CLASS}`)).toHaveLength(1)
  })

  it('does not add preview for invalid word tokens', () => {
    const source = '<span>div</span>{<span>color</span>:<span>skeyblue</span>;}'
    const rendered = colorizeCss(source)
    const out = document.createElement('div')
    out.innerHTML = rendered

    expect(out.querySelectorAll(`.${COLORIZE_CLASS}`)).toHaveLength(0)
  })
})
