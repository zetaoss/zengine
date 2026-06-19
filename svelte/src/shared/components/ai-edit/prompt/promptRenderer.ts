import type { PromptPart, TemplateVariable } from './promptTypes'

export type { PromptPart, TemplateVariable } from './promptTypes'

function canonicalVariableName(name: string) {
  return name.trim()
}

export function isReservedAutoVariableName(rawName: string) {
  const canonical = canonicalVariableName(rawName)
  return canonical === '제목' || canonical === '기존 문서 내용'
}

export function extractTemplateVariables(template: string): TemplateVariable[] {
  const regex = /(```[\s\S]*?```)|(\{([^}]+)\})/g
  const variables = new Map<string, TemplateVariable>()

  let match
  while ((match = regex.exec(template)) !== null) {
    const [, codeBlock, , name] = match
    if (codeBlock) continue

    const variableName = canonicalVariableName(name)
    if (!variableName || isReservedAutoVariableName(variableName)) continue

    variables.set(variableName, {
      name: variableName,
    })
  }

  return Array.from(variables.values())
}

export function usesTemplateVariable(template: string, variableName: string): boolean {
  const regex = /\{([^}]+)\}/g
  for (const match of template.matchAll(regex)) {
    if (canonicalVariableName(match[1]) === variableName) return true
  }
  return false
}

function resolvePlaceholderValue({
  variableNameRaw,
  displayTitle,
  existingContent,
  variableValues,
}: {
  variableNameRaw: string
  displayTitle: string
  existingContent: string
  variableValues: Record<string, string>
}) {
  const variableName = canonicalVariableName(variableNameRaw)

  if (variableName === '제목') {
    return displayTitle.trim()
  }
  if (variableName === '기존 문서 내용') {
    const editedContent = variableValues[variableName]
    return typeof editedContent === 'string' ? editedContent : existingContent
  }

  return (variableValues[variableName] || '').trim()
}

/**
 * Core rendering logic that unifies preview and final prompt.
 * Now optimized for Live Editor: preserves literal structure for template reconstruction.
 */
export function getPromptParts({
  template,
  displayTitle,
  existingContent,
  variableValues,
  rawMode = false,
}: {
  template: string
  displayTitle: string
  existingContent: string
  variableValues: Record<string, string>
  rawMode?: boolean
}): PromptPart[] {
  const parts: PromptPart[] = []
  const regex = /(```[\s\S]*?```)|(\{([^}]+)\})/g

  let lastIndex = 0
  let match
  let customOrdinal = 0

  while ((match = regex.exec(template)) !== null) {
    const [fullMatch, codeBlock, , name] = match
    const offset = match.index

    const leadingText = template.substring(lastIndex, offset)
    if (leadingText) {
      parts.push({ text: leadingText, type: 'plain', templateStart: lastIndex, templateEnd: offset })
    }

    if (codeBlock) {
      const processedCode = rawMode
        ? codeBlock
        : codeBlock.replace(/\{([^}]+)\}/g, (innerMatch, innerName) => {
            if (!isReservedAutoVariableName(innerName)) return innerMatch
            return resolvePlaceholderValue({
              variableNameRaw: innerName,
              displayTitle,
              existingContent,
              variableValues,
            })
          })

      const firstLine = processedCode.split('\n', 1)[0] ?? ''
      const lang = firstLine.replace(/^```/, '').trim().toLowerCase()
      parts.push({
        text: processedCode,
        type: 'code',
        lang: lang || undefined,
        templateText: codeBlock,
        templateStart: offset,
        templateEnd: regex.lastIndex,
      })
    } else {
      const label = canonicalVariableName(name)
      const val = resolvePlaceholderValue({ variableNameRaw: name, displayTitle, existingContent, variableValues })
      const source: NonNullable<PromptPart['source']> = isReservedAutoVariableName(label) ? 'preset' : 'custom'

      const textBefore = template.substring(0, offset)
      const textAfter = template.substring(offset + fullMatch.length)
      const isLineStart = textBefore === '' || textBefore.endsWith('\n')
      const isLineEnd = textAfter === '' || textAfter.startsWith('\n')
      const lastNewlineIndex = textBefore.lastIndexOf('\n')
      const nextNewlineIndex = textAfter.indexOf('\n')
      const lineBefore = lastNewlineIndex === -1 ? textBefore : textBefore.substring(lastNewlineIndex + 1)
      const lineAfter = nextNewlineIndex === -1 ? textAfter : textAfter.substring(0, nextNewlineIndex)
      const isInlineOnlyVariable = label === '제목'
      const isBlock = !isInlineOnlyVariable && isLineStart && isLineEnd && lineBefore.trim() === '' && lineAfter.trim() === ''

      parts.push({
        text: rawMode ? `{${name}}` : val,
        type: isBlock ? 'block' : 'inline',
        source,
        label,
        ordinal: source === 'custom' ? ++customOrdinal : undefined,
        templateText: fullMatch,
        templateStart: offset,
        templateEnd: regex.lastIndex,
      })
    }

    lastIndex = regex.lastIndex
  }

  const trailingText = template.substring(lastIndex)
  if (trailingText) {
    parts.push({ text: trailingText, type: 'plain', templateStart: lastIndex, templateEnd: template.length })
  }

  return parts
}

export function renderFinalPrompt(args: {
  template: string
  displayTitle: string
  existingContent: string
  variableValues: Record<string, string>
}) {
  const parts = getPromptParts(args)
  let output = ''
  let suppressNextLeadingNewline = false

  for (const part of parts) {
    if (part.type === 'block' && !part.text.trim()) {
      suppressNextLeadingNewline = true
      continue
    }

    let text = part.text
    if (suppressNextLeadingNewline) {
      text = text.replace(/^\n/, '')
      suppressNextLeadingNewline = false
    }

    if (part.type === 'block' && part.label) {
      const trimmedText = text.trim()
      if (trimmedText) {
        output += `# ${part.label}\n${trimmedText}`
      }
    } else {
      output += text
    }
  }

  const trimmedOutput = output.trim()
  return trimmedOutput ? `${trimmedOutput}\n` : ''
}
