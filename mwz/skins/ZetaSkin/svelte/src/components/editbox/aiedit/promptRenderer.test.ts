import { describe, expect, it } from 'vitest'

import { getPromptParts, renderFinalPrompt } from './promptRenderer'

describe('renderFinalPrompt', () => {
  const defaultArgs = {
    template: '',
    displayTitle: 'Test Title',
    existingContent: 'Existing Content',
    customFieldValues: {},
  }

  it('should substitute automatic and custom variables with correct formatting', () => {
    const template = '{제목}\n{기존 문서 내용}\n{# 추가정보}'
    const customFieldValues = { 추가정보: '정보내용' }

    const result = renderFinalPrompt({ ...defaultArgs, template, customFieldValues })

    // Checks:
    // 1. {제목} -> Title
    // 2. {기존 문서 내용} -> Existing
    // 3. {# 추가정보} -> # 추가정보\nValue
    // 4. Normalization (gap before header)
    expect(result).toBe('Test Title\n```text\nExisting Content\n```\n\n# 추가정보\n정보내용')
  })

  it('should normalize header spacing (blank line before, no blank line after)', () => {
    const template = 'Text\n\n\n\n# Header1\n\n\nContent\n\n\n## Header2'
    const result = renderFinalPrompt({ ...defaultArgs, template })

    // Expected:
    // - Collapsed 4 newlines to 2 before Header1
    // - Removed blank lines after Header1
    // - Collapsed 3 newlines to 2 before Header2
    expect(result).toBe('Text\n\n# Header1\nContent\n\n## Header2')
  })

  it('should handle variable states (empty vs non-empty) and surrounding gaps', () => {
    // Case 1: Empty variable between headers
    const template = '# Header1\n\n{# Foo}\n\n# Header2'
    expect(renderFinalPrompt({ ...defaultArgs, template, customFieldValues: { Foo: '' } }))
      .toBe('# Header1\n\n# Header2')

    // Case 2: Non-empty variable between headers
    expect(renderFinalPrompt({ ...defaultArgs, template, customFieldValues: { Foo: 'Bar' } }))
      .toBe('# Header1\n\n# Foo\nBar\n\n# Header2')
  })

  it('should render {# name:code} as a fenced code block', () => {
    const template = '{# foo:code}'
    const result = renderFinalPrompt({ ...defaultArgs, template, customFieldValues: { foo: 'bar' } })

    expect(result).toBe('# foo\n```\nbar\n```')
  })

  it('should render {# name:code:lang} as a fenced code block with language', () => {
    const template = '{# foo:code:go}'
    const result = renderFinalPrompt({ ...defaultArgs, template, customFieldValues: { foo: 'bar' } })

    expect(result).toBe('# foo\n```go\nbar\n```')
  })

  it('should classify ```text fenced blocks as text parts', () => {
    const template = 'before\n```text\nhello world\n```\nafter'
    const parts = getPromptParts({ ...defaultArgs, template })

    expect(parts).toEqual([
      { text: 'before\n', type: 'plain' },
      { text: '```text\nhello world\n```', type: 'text' },
      { text: '\nafter', type: 'plain' },
    ])
  })

  it('should render edit auto variables and inline code rendering', () => {
    const template = '{제목}\n\n{기존 문서 내용:code}'
    const result = renderFinalPrompt({
      ...defaultArgs,
      displayTitle: '표시 제목 샘플',
      existingContent: '기존 문서 본문',
      template,
    })

    expect(result).toBe('표시 제목 샘플\n\n```\n기존 문서 본문\n```')
  })

  it('should not resolve legacy aliases for auto variables', () => {
    const template = '{제목}\n{표시제목}\n{기존문서}\n{문서내용}'
    const result = renderFinalPrompt({
      ...defaultArgs,
      displayTitle: '동일 제목',
      existingContent: '동일 본문',
      template,
      customFieldValues: { 표시제목: '', 기존문서: '', 문서내용: '' },
    })

    expect(result).toBe('동일 제목')
  })

  it('should render reserved auto variables inside fenced code blocks', () => {
    const template = '```text\n{기존 문서 내용}\n```'
    const result = renderFinalPrompt({
      ...defaultArgs,
      existingContent: '치환된 본문',
      template,
    })

    expect(result).toBe('```text\n치환된 본문\n```')
  })

  it('should render {기존 문서 내용} as fenced text block by default', () => {
    const template = '{기존 문서 내용}'
    const result = renderFinalPrompt({
      ...defaultArgs,
      existingContent: '치환된 본문',
      template,
    })

    expect(result).toBe('```text\n치환된 본문\n```')
  })

  it('should append additional content after the rendered prompt', () => {
    const template = '# Header'
    const result = renderFinalPrompt({
      ...defaultArgs,
      template,
      notes: '추가 메모',
    })

    expect(result).toBe('# Header\n\n추가 메모')
  })

  it('should include additional content in preview parts', () => {
    const template = '# Header'
    const parts = getPromptParts({
      ...defaultArgs,
      template,
      notes: '추가 메모',
    })

    expect(parts).toEqual([
      { text: '# Header', type: 'plain' },
      { text: '\n\n', type: 'plain' },
      { text: '추가 메모', type: 'block', source: 'preset' },
    ])
  })

  it('should protect code blocks and ignore headers inside them', () => {
    const template = '# Header\n\n```text\n# Not a header\n\n\nInside code\n```\n{# Var}'
    const result = renderFinalPrompt({ ...defaultArgs, template, customFieldValues: { Var: 'Val' } })

    expect(result).toBe('# Header\n```text\n# Not a header\n\n\nInside code\n```\n\n# Var\nVal')
  })

  it('should handle the user reported wikitext template structure', () => {
    const template = `# 문서 형식

\`\`\`wikitext
==개요==
[[분류:적절한 분류]]
\`\`\`

{# 추가정보}

# 출력`
    
    const result = renderFinalPrompt({ ...defaultArgs, template, customFieldValues: { 추가정보: '' } })

    const expected = `# 문서 형식
\`\`\`wikitext
==개요==
[[분류:적절한 분류]]
\`\`\`

# 출력`
    expect(result).toBe(expected)
  })

  it('should preserve empty block placeholders in preview mode when requested', () => {
    const template = '# Header1\n\n{# Foo}\n\n{Bar}\n\n# Header2'

    const parts = getPromptParts({
      ...defaultArgs,
      template,
      customFieldValues: { Foo: '', Bar: '' },
      preserveEmptyBlockPlaceholders: true,
    })

    expect(parts).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ text: '', type: 'block', label: 'Foo' }),
        expect.objectContaining({ text: '', type: 'inline', label: 'Bar' }),
      ]),
    )
  })
})
