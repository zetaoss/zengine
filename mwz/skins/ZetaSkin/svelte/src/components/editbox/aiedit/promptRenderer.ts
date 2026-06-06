export interface TemplateVariable {
  name: string
  hashes: string
}

export interface PromptPart {
  text: string
  type: 'plain' | 'inline' | 'block' | 'code' | 'text'
  source?: 'preset' | 'custom'
  ordinal?: number
  label?: string
}

function parsePlaceholderSpec(rawName: string) {
  const [namePart, ...modeParts] = rawName.split(':')
  const blockMode = (modeParts[0] ?? '').trim().toLowerCase()
  const blockLang = (modeParts[1] ?? '').trim()
  return {
    name: namePart.trim(),
    blockMode,
    blockLang,
  }
}

function canonicalVariableName(name: string) {
  return name.trim()
}

function isPresetVariableName(rawName: string) {
  const canonical = canonicalVariableName(rawName)
  return canonical === '제목' || canonical === '기존 문서 내용'
}

function isReservedAutoVariableName(rawName: string) {
  const canonical = canonicalVariableName(rawName)
  return canonical === '제목' || canonical === '기존 문서 내용'
}

function toCodeFence(value: string, lang: string) {
  const codeFence = lang ? `\`\`\`${lang}` : '```'
  return `${codeFence}\n${value}\n\`\`\``
}

function getFencedBlockLanguage(text: string) {
  const firstLine = text.split('\n', 1)[0] ?? ''
  return firstLine.replace(/^```/, '').trim().toLowerCase()
}

function resolvePlaceholderValue({
  variableNameRaw,
  hashes,
  blockMode,
  blockLang,
  displayTitle,
  existingContent,
  customFieldValues,
}: {
  variableNameRaw: string
  hashes: string
  blockMode: string
  blockLang: string
  displayTitle: string
  existingContent: string
  customFieldValues: Record<string, string>
}) {
  const variableName = canonicalVariableName(variableNameRaw)
  let val: string

  if (variableName === '제목') {
    val = displayTitle.trim()
  } else if (variableName === '기존 문서 내용') {
    const editedContent = customFieldValues[variableName]
    val = typeof editedContent === 'string' ? editedContent : existingContent
  } else {
    const fieldVal = (customFieldValues[variableName] || '').trim()
    if (!fieldVal) return ''
    val = fieldVal
  }

  if (!val) return ''

  if (hashes) {
    if (blockMode === 'code') {
      return `${hashes} ${variableName}\n${toCodeFence(val, blockLang)}\n`
    }
    return `${hashes} ${variableName}\n${val}\n`
  }

  if (blockMode === 'code') return toCodeFence(val, blockLang)
  return val
}

/**
 * Core rendering logic that unifies preview and final prompt.
 */
export function getPromptParts({
  template,
  displayTitle,
  existingContent,
  customFieldValues,
  notes = '',
  rawMode = false,
  preserveEmptyBlockPlaceholders = false,
}: {
  template: string
  displayTitle: string
  existingContent: string
  customFieldValues: Record<string, string>
  notes?: string
  rawMode?: boolean
  preserveEmptyBlockPlaceholders?: boolean
}): PromptPart[] {
  const regex = /(```[\s\S]*?```)|(\{((?:#+\s*)?)([^}]+)\})/g
  const placeholderMeta: Array<{ label: string }> = []

  // Stage 1: Replace variables with markers wrapping their values
  let out = template.replace(regex, (match, codeBlock, fullPlaceholder, hashes, name) => {
    if (codeBlock) {
      if (rawMode) return codeBlock
      return codeBlock.replace(/\{((?:#+\s*)?)([^}]+)\}/g, (innerMatch: string, _innerHashes: string, innerName: string) => {
        const parsed = parsePlaceholderSpec(innerName)
        if (!isReservedAutoVariableName(parsed.name)) return innerMatch
        return resolvePlaceholderValue({
          variableNameRaw: parsed.name,
          hashes: '',
          blockMode: parsed.blockMode,
          blockLang: parsed.blockLang,
          displayTitle,
          existingContent,
          customFieldValues,
        })
      })
    }

    if (rawMode) {
      const type: PromptPart['type'] = hashes ? 'block' : 'inline'
      const typeChar = type === 'inline' ? 'I' : 'B'
      const sourceChar = isPresetVariableName(name) ? 'P' : 'C'
      return `\uE000${typeChar}${sourceChar}${fullPlaceholder}\uE001`
    }

    const { name: variableNameRaw, blockMode, blockLang } = parsePlaceholderSpec(name)
    hashes = (hashes || '').trim()
    let type: PromptPart['type'] = hashes ? 'block' : 'inline'
    const variableName = canonicalVariableName(variableNameRaw)
    const source: NonNullable<PromptPart['source']> = isPresetVariableName(variableNameRaw) ? 'preset' : 'custom'
    const label = variableName
    const isAutoFencedDocumentContent = variableName === '기존 문서 내용'
    const val = resolvePlaceholderValue({
      variableNameRaw,
      hashes,
      blockMode,
      blockLang,
      displayTitle,
      existingContent,
      customFieldValues,
    })
    if (!val) {
      if (preserveEmptyBlockPlaceholders) {
        if (blockMode === 'code' || (isAutoFencedDocumentContent && !blockMode)) {
          type = 'block'
        }
        const typeChar = type === 'inline' ? 'I' : 'B'
        const sourceChar = source === 'preset' ? 'P' : 'C'
        placeholderMeta.push({ label })
        return `\uE000${typeChar}${sourceChar}\uE001`
      }
      return ''
    }
    if (hashes || blockMode === 'code' || (isAutoFencedDocumentContent && !blockMode)) type = 'block'

    if (isAutoFencedDocumentContent && !hashes && !blockMode) {
      const normalized = val.trimEnd()
      const fenced = toCodeFence(normalized, 'text')
      const typeChar = 'B'
      placeholderMeta.push({ label })
      return `\uE000${typeChar}P${fenced}\uE001`
    }

    const typeChar = type === 'inline' ? 'I' : 'B'
    const sourceChar = source === 'preset' ? 'P' : 'C'
    placeholderMeta.push({ label })
    return `\uE000${typeChar}${sourceChar}${val}\uE001`
  })

  // Stage 2: Normalization (only if not in raw mode)
  if (!rawMode) {
    out = out
      .split('\n')
      .map((line) => line.trimEnd())
      .join('\n')

    const codeBlocks: string[] = []
    out = out.replace(/```[\s\S]*?```/g, (match) => {
      codeBlocks.push(match)
      return `__CODE_BLOCK_${codeBlocks.length - 1}__`
    })

    // 1. Remove blank lines after headers
    out = out.replace(/(#+\s[^\n]*\n)([\s\uE000\uE001IBPC]+)/g, (match, header, trailing) => {
      const preserved = trailing.replace(/[^\uE000\uE001IBPC]/g, '')
      return header + preserved
    })

    // 2. Ensure exactly one blank line before headers
    out = out.replace(/([\s\uE000\uE001IBPC]+)(#+\s)/g, (match, leading, header) => {
      const preserved = leading.replace(/[^\uE000\uE001IBPC]/g, '')
      return '\n\n' + preserved + header
    })

    out = out.replace(/__CODE_BLOCK_(\d+)__/g, (_, index) => codeBlocks[Number(index)])
    out = out.trim()
  }

  // Stage 3: Split into variable/plain parts
  const result: PromptPart[] = []
  const tokens = out.split(/(\uE000[IB][PC][^\uE001]*\uE001)/g)
  let customOrdinal = 0

  for (const token of tokens) {
    if (token.startsWith('\uE000') && token.endsWith('\uE001')) {
      const typeChar = token.charAt(1)
      const sourceChar = token.charAt(2)
      const text = token.substring(3, token.length - 1)
      const source = sourceChar === 'P' ? 'preset' : 'custom'
      const { label = '' } = placeholderMeta.shift() ?? {}
      result.push({
        text,
        type: typeChar === 'I' ? 'inline' : 'block',
        source,
        ordinal: source === 'custom' ? ++customOrdinal : undefined,
        label,
      })
    } else if (token) {
      result.push({ text: token, type: 'plain' })
    }
  }

  // Stage 4: Split plain parts by fenced code blocks
  const withCodeParts: PromptPart[] = []
  for (const part of result) {
    if (part.type !== 'plain') {
      withCodeParts.push(part)
      continue
    }

    const codeTokens = part.text.split(/(```[\s\S]*?```)/g)
    for (const token of codeTokens) {
      if (!token) continue
      const isFencedBlock = token.startsWith('```') && token.endsWith('```')
      withCodeParts.push({
        text: token,
        type: isFencedBlock ? (getFencedBlockLanguage(token) === 'text' ? 'text' : 'code') : 'plain',
      })
    }
  }

  const filteredParts = withCodeParts.filter((p) => p.text !== '')
  const trimmedAdditionalContent = notes.trim()
  const outputParts = preserveEmptyBlockPlaceholders ? withCodeParts : filteredParts
  if (!trimmedAdditionalContent) return outputParts

  return [
    ...outputParts,
    { text: '\n\n', type: 'plain' },
    {
      text: trimmedAdditionalContent,
      type: 'block',
      source: 'preset',
    },
  ]
}

export function renderFinalPrompt(args: {
  template: string
  displayTitle: string
  existingContent: string
  templateVariables?: TemplateVariable[]
  customFieldValues: Record<string, string>
  notes?: string
}) {
  const parts = getPromptParts(args)
  return parts
    .map((p) => p.text)
    .join('')
    .trim()
}
