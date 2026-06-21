export type DiffKind = 'context' | 'added' | 'removed'

export interface DiffLine {
  kind: DiffKind
  newLine?: number
  oldLine?: number
  text: string
}

export interface InlineSegment {
  changed: boolean
  text: string
}

export interface SplitDiffRow {
  changed: boolean
  newInline?: InlineSegment[]
  newLine?: DiffLine
  oldInline?: InlineSegment[]
  oldLine?: DiffLine
}

const maxInlineMatrixCells = 40_000
const maxMatrixCells = 2_000_000

export function createDiff(original: string, revised: string): DiffLine[] {
  const oldLines = original.split('\n')
  const newLines = revised.split('\n')
  const oldLength = oldLines.length
  const newLength = newLines.length

  if (oldLength * newLength > maxMatrixCells) {
    return createFallbackDiff(oldLines, newLines)
  }

  const width = newLength + 1
  const matrix = new Uint32Array((oldLength + 1) * width)
  for (let oldIndex = oldLength - 1; oldIndex >= 0; oldIndex -= 1) {
    for (let newIndex = newLength - 1; newIndex >= 0; newIndex -= 1) {
      const index = oldIndex * width + newIndex
      matrix[index] =
        oldLines[oldIndex] === newLines[newIndex]
          ? matrix[(oldIndex + 1) * width + newIndex + 1] + 1
          : Math.max(matrix[(oldIndex + 1) * width + newIndex], matrix[oldIndex * width + newIndex + 1])
    }
  }

  const lines: DiffLine[] = []
  let oldIndex = 0
  let newIndex = 0
  while (oldIndex < oldLength && newIndex < newLength) {
    if (oldLines[oldIndex] === newLines[newIndex]) {
      lines.push({ kind: 'context', oldLine: oldIndex + 1, newLine: newIndex + 1, text: oldLines[oldIndex] })
      oldIndex += 1
      newIndex += 1
    } else if (matrix[(oldIndex + 1) * width + newIndex] >= matrix[oldIndex * width + newIndex + 1]) {
      lines.push({ kind: 'removed', oldLine: oldIndex + 1, text: oldLines[oldIndex] })
      oldIndex += 1
    } else {
      lines.push({ kind: 'added', newLine: newIndex + 1, text: newLines[newIndex] })
      newIndex += 1
    }
  }
  while (oldIndex < oldLength) {
    lines.push({ kind: 'removed', oldLine: oldIndex + 1, text: oldLines[oldIndex] })
    oldIndex += 1
  }
  while (newIndex < newLength) {
    lines.push({ kind: 'added', newLine: newIndex + 1, text: newLines[newIndex] })
    newIndex += 1
  }
  return lines
}

function createFallbackDiff(oldLines: string[], newLines: string[]): DiffLine[] {
  let prefixLength = 0
  while (prefixLength < oldLines.length && prefixLength < newLines.length && oldLines[prefixLength] === newLines[prefixLength]) {
    prefixLength += 1
  }

  let suffixLength = 0
  while (
    suffixLength < oldLines.length - prefixLength &&
    suffixLength < newLines.length - prefixLength &&
    oldLines[oldLines.length - suffixLength - 1] === newLines[newLines.length - suffixLength - 1]
  ) {
    suffixLength += 1
  }

  const lines: DiffLine[] = []
  for (let index = 0; index < prefixLength; index += 1) {
    lines.push({ kind: 'context', oldLine: index + 1, newLine: index + 1, text: oldLines[index] })
  }
  for (let index = prefixLength; index < oldLines.length - suffixLength; index += 1) {
    lines.push({ kind: 'removed', oldLine: index + 1, text: oldLines[index] })
  }
  for (let index = prefixLength; index < newLines.length - suffixLength; index += 1) {
    lines.push({ kind: 'added', newLine: index + 1, text: newLines[index] })
  }
  for (let offset = suffixLength; offset > 0; offset -= 1) {
    const oldIndex = oldLines.length - offset
    const newIndex = newLines.length - offset
    lines.push({ kind: 'context', oldLine: oldIndex + 1, newLine: newIndex + 1, text: oldLines[oldIndex] })
  }
  return lines
}

export function createSplitRows(lines: DiffLine[]): SplitDiffRow[] {
  const rows: SplitDiffRow[] = []
  let index = 0
  while (index < lines.length) {
    const line = lines[index]
    if (line.kind === 'context') {
      rows.push({ changed: false, oldLine: line, newLine: line })
      index += 1
      continue
    }

    const removed: DiffLine[] = []
    const added: DiffLine[] = []
    while (index < lines.length && lines[index].kind !== 'context') {
      const changedLine = lines[index]
      if (changedLine.kind === 'removed') removed.push(changedLine)
      else added.push(changedLine)
      index += 1
    }
    const rowCount = Math.max(removed.length, added.length)
    for (let rowIndex = 0; rowIndex < rowCount; rowIndex += 1) {
      const oldLine = removed[rowIndex]
      const newLine = added[rowIndex]
      const inlineDiff = oldLine && newLine ? createInlineDiff(oldLine.text, newLine.text) : undefined
      rows.push({
        changed: true,
        oldLine,
        newLine,
        oldInline: inlineDiff?.oldSegments,
        newInline: inlineDiff?.newSegments,
      })
    }
  }
  return rows
}

function createInlineDiff(oldText: string, newText: string) {
  const oldTokens = tokenize(oldText)
  const newTokens = tokenize(newText)
  if (oldTokens.length * newTokens.length > maxInlineMatrixCells) {
    return {
      oldSegments: [{ changed: true, text: oldText }],
      newSegments: [{ changed: true, text: newText }],
    }
  }

  const width = newTokens.length + 1
  const matrix = new Uint16Array((oldTokens.length + 1) * width)
  for (let oldIndex = oldTokens.length - 1; oldIndex >= 0; oldIndex -= 1) {
    for (let newIndex = newTokens.length - 1; newIndex >= 0; newIndex -= 1) {
      const index = oldIndex * width + newIndex
      matrix[index] =
        oldTokens[oldIndex] === newTokens[newIndex]
          ? matrix[(oldIndex + 1) * width + newIndex + 1] + 1
          : Math.max(matrix[(oldIndex + 1) * width + newIndex], matrix[oldIndex * width + newIndex + 1])
    }
  }

  const oldSegments: InlineSegment[] = []
  const newSegments: InlineSegment[] = []
  let oldIndex = 0
  let newIndex = 0
  while (oldIndex < oldTokens.length && newIndex < newTokens.length) {
    if (oldTokens[oldIndex] === newTokens[newIndex]) {
      appendInlineSegment(oldSegments, oldTokens[oldIndex], false)
      appendInlineSegment(newSegments, newTokens[newIndex], false)
      oldIndex += 1
      newIndex += 1
    } else if (matrix[(oldIndex + 1) * width + newIndex] >= matrix[oldIndex * width + newIndex + 1]) {
      appendInlineSegment(oldSegments, oldTokens[oldIndex], true)
      oldIndex += 1
    } else {
      appendInlineSegment(newSegments, newTokens[newIndex], true)
      newIndex += 1
    }
  }
  while (oldIndex < oldTokens.length) {
    appendInlineSegment(oldSegments, oldTokens[oldIndex], true)
    oldIndex += 1
  }
  while (newIndex < newTokens.length) {
    appendInlineSegment(newSegments, newTokens[newIndex], true)
    newIndex += 1
  }
  return { oldSegments, newSegments }
}

function tokenize(text: string) {
  return text.match(/\s+|[\p{L}\p{N}_]+|[^\s\p{L}\p{N}_]+/gu) ?? []
}

function appendInlineSegment(segments: InlineSegment[], text: string, changed: boolean) {
  const previous = segments.at(-1)
  if (previous?.changed === changed) {
    previous.text += text
    return
  }
  segments.push({ changed, text })
}

export function compactDiff(rows: SplitDiffRow[], context: number): Array<SplitDiffRow | null> {
  const changedIndexes = rows.flatMap((row, index) => (row.changed ? [index] : []))
  if (!changedIndexes.length) return []

  const ranges: Array<{ end: number; start: number }> = []
  for (const index of changedIndexes) {
    const start = Math.max(0, index - context)
    const end = Math.min(rows.length - 1, index + context)
    const previous = ranges.at(-1)
    if (previous && start <= previous.end + 1) {
      previous.end = Math.max(previous.end, end)
    } else {
      ranges.push({ start, end })
    }
  }

  return ranges.flatMap((range, index) => [...(index > 0 ? [null] : []), ...rows.slice(range.start, range.end + 1)])
}
