import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { linkifyWikiMatch } from '@/utils/linkify'

describe('linkifyWikiMatch', () => {
  let errorSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  afterEach(() => {
    errorSpy.mockRestore()
    vi.clearAllMocks()
  })

  it('should convert an existing wiki link', async () => {
    const result = await linkifyWikiMatch(
      'This is a [[TestPage]] link',
      '[[TestPage]]',
      { TestPage: true },
    )
    expect(result).toBe(
      'This is a <a href="/wiki/TestPage" class="internal">TestPage</a> link',
    )
  })

  it('should convert a non-existing wiki link', async () => {
    const result = await linkifyWikiMatch(
      'This is a [[MissingPage]] link',
      '[[MissingPage]]',
      { MissingPage: false },
    )
    expect(result).toBe(
      'This is a <a href="/wiki/MissingPage" class="internal new">MissingPage</a> link',
    )
  })

  it('should handle multiple same links', async () => {
    const result = await linkifyWikiMatch(
      '[[TestPage]] is linked twice: [[TestPage]].',
      '[[TestPage]]',
      { TestPage: true },
    )
    expect(result).toBe(
      '<a href="/wiki/TestPage" class="internal">TestPage</a> is linked twice: <a href="/wiki/TestPage" class="internal">TestPage</a>.',
    )
  })

  it('should handle multiple different links (1)', async () => {
    const result = await linkifyWikiMatch(
      'Links: [[TestPage1]], [[TestPage2]].',
      '[[TestPage1]]',
      { TestPage1: true },
    )
    expect(result).toBe(
      'Links: <a href="/wiki/TestPage1" class="internal">TestPage1</a>, [[TestPage2]].',
    )
  })

  it('should handle multiple different links (2)', async () => {
    const result = await linkifyWikiMatch(
      'Links: [[TestPage1]], [[TestPage2]].',
      '[[TestPage2]]',
      { TestPage2: true },
    )
    expect(result).toBe(
      'Links: [[TestPage1]], <a href="/wiki/TestPage2" class="internal">TestPage2</a>.',
    )
  })

  it('should handle multiple different links (3)', async () => {
    const result = await linkifyWikiMatch(
      'Links: <a href="/wiki/TestPage1" class="internal">TestPage1</a>, [[TestPage2]].',
      '[[TestPage2]]',
      { TestPage2: true },
    )
    expect(result).toBe(
      'Links: <a href="/wiki/TestPage1" class="internal">TestPage1</a>, <a href="/wiki/TestPage2" class="internal">TestPage2</a>.',
    )
  })

  it('should fallback on API failure', async () => {
    const result = await linkifyWikiMatch(
      'This is a [[ErrorPage]] link',
      '[[ErrorPage]]',
      { ErrorPage: false },
    )
    expect(result).toBe(
      'This is a <a href="/wiki/ErrorPage" class="internal new">ErrorPage</a> link',
    )
  })
})
