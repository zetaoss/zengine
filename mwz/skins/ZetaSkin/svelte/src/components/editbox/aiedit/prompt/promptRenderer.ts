function canonicalVariableName(name: string) {
  return name.trim()
}

export function isReservedAutoVariableName(rawName: string) {
  const canonical = canonicalVariableName(rawName)
  return canonical === '제목' || canonical === '기존 문서 내용'
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
  renderExistingContent,
}: {
  variableNameRaw: string
  displayTitle: string
  existingContent: string
  variableValues: Record<string, string>
  renderExistingContent: boolean
}) {
  const variableName = canonicalVariableName(variableNameRaw)

  if (variableName === '제목') {
    return displayTitle.trim()
  }
  if (variableName === '기존 문서 내용') {
    if (!renderExistingContent) return `{${variableNameRaw}}`

    const editedContent = variableValues[variableName]
    return typeof editedContent === 'string' ? editedContent : existingContent
  }

  return `{${variableNameRaw}}`
}

function renderPrompt({
  template,
  displayTitle,
  existingContent,
  variableValues,
  renderExistingContent = true,
}: {
  template: string
  displayTitle: string
  existingContent: string
  variableValues: Record<string, string>
  renderExistingContent?: boolean
}) {
  let output = ''
  const regex = /(```[\s\S]*?```)|(\{([^}]+)\})/g

  let lastIndex = 0
  let match
  while ((match = regex.exec(template)) !== null) {
    const [fullMatch, codeBlock, , name] = match
    const offset = match.index

    output += template.substring(lastIndex, offset)

    if (codeBlock) {
      output += codeBlock.replace(/\{([^}]+)\}/g, (innerMatch, innerName) => {
        if (!isReservedAutoVariableName(innerName)) return innerMatch
        return resolvePlaceholderValue({
          variableNameRaw: innerName,
          displayTitle,
          existingContent,
          variableValues,
          renderExistingContent,
        })
      })
    } else {
      const label = canonicalVariableName(name)
      if (!isReservedAutoVariableName(label)) {
        output += fullMatch
        lastIndex = regex.lastIndex
        continue
      }

      output += resolvePlaceholderValue({ variableNameRaw: name, displayTitle, existingContent, variableValues, renderExistingContent })
    }

    lastIndex = regex.lastIndex
  }

  return output + template.substring(lastIndex)
}

export function renderFinalPrompt(args: {
  template: string
  displayTitle: string
  existingContent: string
  variableValues: Record<string, string>
}) {
  const output = renderPrompt({ ...args, renderExistingContent: true })

  const trimmedOutput = output.trim()
  return trimmedOutput ? `${trimmedOutput}\n` : ''
}
