import { describe, expect, it } from 'vitest'

import { extractTemplateVariables, getPromptParts, renderFinalPrompt, usesTemplateVariable } from './promptRenderer'

describe('renderFinalPrompt', () => {
  const defaultArgs = {
    template: '',
    displayTitle: 'Test Title',
    existingContent: 'Existing Content',
    variableValues: {},
  }

  it('should substitute variables and add headers for blocks', () => {
    const template = '{제목}\n\n# Manual Header\n```text\n{기존 문서 내용}\n```\n\n{추가정보}'
    const variableValues = { 추가정보: '정보내용' }

    const result = renderFinalPrompt({ ...defaultArgs, template, variableValues })

    // {기존 문서 내용} is inline (inside code block) -> Existing Content
    // {추가정보} is block -> # 추가정보\n정보내용
    expect(result).toContain('Test Title\n\n# Manual Header')
    expect(result).toContain('# Manual Header\n```text\nExisting Content\n```')
    expect(result).toContain('# 추가정보\n정보내용')
  })

  it('should add a header for block variables with spaced labels', () => {
    const template = '{추가 정보}'
    const variableValues = { '추가 정보': '사용자가 입력한 내용' }

    const result = renderFinalPrompt({ ...defaultArgs, template, variableValues })

    expect(result).toBe('# 추가 정보\n사용자가 입력한 내용\n')
  })

  it('should NOT re-render resolved content as template', () => {
    const template = '{기존 문서 내용}'
    const existingContent = 'Hello {제목}'

    const result = renderFinalPrompt({ ...defaultArgs, template, existingContent })

    // Content should remain literal, even if it looks like a variable
    expect(result).toContain('Hello {제목}')
  })

  it('should remove entire block variable section (including auto-header) if empty', () => {
    const template = 'Keep\n{EmptyBlock}\nKeep'
    const result = renderFinalPrompt({ ...defaultArgs, template, variableValues: { EmptyBlock: '' } })

    expect(result).toBe('Keep\nKeep\n')
  })

  it('should handle reserved variables correctly', () => {
    const template = 'Title: {제목}' // inline
    const result = renderFinalPrompt({ ...defaultArgs, template, displayTitle: 'My Title' })

    expect(result).toBe('Title: My Title\n')
  })
})

describe('getPromptParts', () => {
  const defaultArgs = {
    template: '',
    displayTitle: '문서 제목',
    existingContent: '본문',
    variableValues: {},
  }

  it('should preserve code block template text while rendering reserved values', () => {
    const template = 'Before\n```mediawiki\n{제목}\n{사용자변수}\n```\nAfter'

    const parts = getPromptParts({ ...defaultArgs, template })
    const codePart = parts.find((part) => part.type === 'code')

    expect(codePart?.text).toBe('```mediawiki\n문서 제목\n{사용자변수}\n```')
    expect(codePart?.templateText).toBe('```mediawiki\n{제목}\n{사용자변수}\n```')
  })

  it('should expose template offsets for preview insertion controls', () => {
    const template = 'Before\n\n{추가 정보}\nAfter'

    const parts = getPromptParts({ ...defaultArgs, template })

    expect(parts[0]).toMatchObject({ type: 'plain', text: 'Before\n\n', templateStart: 0, templateEnd: 8 })
    expect(parts[1]).toMatchObject({ type: 'block', label: '추가 정보', templateStart: 8, templateEnd: 15 })
    expect(parts[2]).toMatchObject({ type: 'plain', text: '\nAfter', templateStart: 15, templateEnd: 21 })
  })

  it('should keep title variables inline even when they are on their own line', () => {
    const parts = getPromptParts({ ...defaultArgs, template: '{제목}' })

    expect(parts[0]).toMatchObject({ type: 'inline', label: '제목', text: '문서 제목' })
  })
})

describe('template variables', () => {
  it('should extract only custom variables outside code blocks', () => {
    const template = ['{제목}', '{사용자 변수}', '```text', '{코드안변수}', '```', '{기존 문서 내용}'].join('\n')

    expect(extractTemplateVariables(template)).toEqual([{ name: '사용자 변수' }])
  })

  it('should detect reserved variables even inside code blocks', () => {
    const template = '```text\n{기존 문서 내용}\n```'

    expect(usesTemplateVariable(template, '기존 문서 내용')).toBe(true)
  })
})
