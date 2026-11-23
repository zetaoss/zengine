import type { AIEditRequestType } from './aiEditTypes'

const markupChecks = [
  { pattern: /\[\[/g, label: '위키 링크([[…]])' },
  { pattern: /\{\{/g, label: '틀({{…}})' },
  { pattern: /<ref\b/gi, label: '각주(<ref>)' },
] as const

function countMatches(text: string, pattern: RegExp) {
  return Array.from(text.matchAll(pattern)).length
}

function hasCategoryLine(text: string) {
  return /^\s*\[\[\s*(?:Category|분류)\s*:/im.test(text)
}

function hasNewMarkdownHeading(original: string | undefined, output: string) {
  const originalHeadings = new Map<string, number>()
  for (const line of original?.split(/\r?\n/) ?? []) {
    if (!/^[ \t]*#{1,6}[ \t]+.+$/.test(line)) continue
    originalHeadings.set(line, (originalHeadings.get(line) ?? 0) + 1)
  }

  for (const line of output.split(/\r?\n/)) {
    if (!/^[ \t]*#{1,6}[ \t]+.+$/.test(line)) continue
    const count = originalHeadings.get(line) ?? 0
    if (count === 0) return true
    originalHeadings.set(line, count - 1)
  }
  return false
}

export function getAiEditWarnings(
  requestType: AIEditRequestType,
  original: string | undefined,
  output: string,
): string[] {
  const warnings: string[] = []

  if (requestType === 'update' && original !== undefined) {
    for (const check of markupChecks) {
      if (countMatches(output, check.pattern) < countMatches(original, check.pattern)) {
        warnings.push(`원문보다 ${check.label}가 줄었습니다.`)
      }
    }
    if (hasCategoryLine(original) && !hasCategoryLine(output)) {
      warnings.push('원문에 있던 분류 줄이 결과에서 사라졌습니다.')
    }
  }

  if (/\*\*[^*\n]+\*\*/.test(output)) warnings.push('마크다운 굵게 표기(**)가 포함되어 있습니다.')
  if (hasNewMarkdownHeading(original, output)) warnings.push('마크다운 제목(# )이 포함되어 있습니다.')
  if (output.includes('`')) warnings.push('백틱(`)이 포함되어 있습니다.')

  return warnings
}
