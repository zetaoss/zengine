import { describe, expect, it } from 'vitest'

import { getAiEditWarnings } from './aiEditWarnings'

describe('getAiEditWarnings', () => {
  it('warns when wiki markup or an existing category is lost', () => {
    const original = '[[A]] [[B]] {{Infobox}} <ref>cite</ref>\n[[분류:문서]]'
    const output = '[[A]] {{Infobox}}\n문서'

    expect(getAiEditWarnings('update', original, output)).toEqual([
      '원문보다 위키 링크([[…]])가 줄었습니다.',
      '원문보다 각주(<ref>)가 줄었습니다.',
      '원문에 있던 분류 줄이 결과에서 사라졌습니다.',
    ])
  })

  it('warns about markdown artifacts for create requests', () => {
    expect(getAiEditWarnings('create', undefined, '# 제목\n**강조** `코드`')).toEqual([
      '마크다운 굵게 표기(**)가 포함되어 있습니다.',
      '마크다운 제목(# )이 포함되어 있습니다.',
      '백틱(`)이 포함되어 있습니다.',
    ])
  })

  it('does not warn when wiki markup counts are preserved or increased', () => {
    expect(getAiEditWarnings('update', '[[A]] {{T}} <ref>x</ref>', '[[A]][[B]] {{T}}{{U}} <ref>x</ref><ref>y</ref>')).toEqual([])
  })

  it('does not flag a category that remains in the result', () => {
    expect(getAiEditWarnings('update', '[[Category:A]]', 'Text\n[[Category:A]]')).toEqual([])
  })

  it('does not treat preserved MediaWiki ordered and nested bullet lists as Markdown', () => {
    const lists = '# First item\n## Nested item\n** Nested bullet'

    expect(getAiEditWarnings('update', lists, lists)).toEqual([])
  })

  it('warns about new Markdown heading syntax without flagging unchanged list lines', () => {
    expect(getAiEditWarnings('update', '# Existing list item', '# Existing list item\n# New heading')).toEqual([
      '마크다운 제목(# )이 포함되어 있습니다.',
    ])
  })
})
