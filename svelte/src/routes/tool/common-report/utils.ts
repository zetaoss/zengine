import type { Row } from './types'

export function getRatio(row: Row, idx: number) {
  const item = row.items?.[idx]
  if (!item || row.total === 0) return 0
  return item.total / row.total
}

export function getScore(row: Row): number {
  const nums = row.items.map((x) => x.total)
  const total = nums.reduce((acc, cur) => acc + cur, 0)
  if (nums.length < 2 || total === 0) return -1

  const p = nums.map((x) => x / total)
  const p0 = p[0] ?? 0
  const p1 = p[1] ?? 0
  const gap = p0 - p1
  const sizeWeight = Math.log10(total)
  const weightedDominance = (gap * 0.8 + p0 * 0.2) * sizeWeight
  const maxWeight = Math.log10(1000000)

  return Math.floor(Math.max(0, Math.min(weightedDominance / maxWeight, 1)) * 4)
}

function formatDateYMD(dateInput: string): string {
  const date = new Date(dateInput)
  if (Number.isNaN(date.getTime())) return dateInput

  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

export function getWikitextTable(table: HTMLTableElement, id: number, url: string, createdAt: string) {
  let number = 0
  const rows: string[][] = []
  const trs = table.getElementsByTagName('tr')

  for (let i = 0; i < trs.length; i += 1) {
    const tr = trs[i]
    if (!tr) continue
    const tds = tr.getElementsByTagName('td')
    const row: string[] = []
    for (let j = 0; j < tds.length; j += 1) {
      const td = tds[j] as HTMLElement
      row.push(td.innerText)
    }
    rows.push(row)
    number = row.length
  }

  const dt = formatDateYMD(createdAt)
  let text = `{| class='wikitable'
|+ [${url} 통용 보고서 #${id} ( ${dt} )]
|-
! 표기 !! 비율 !! 계 !! D블로그 !! N블로그 !! N책 !! N뉴스 !! 구글
`

  for (let i = 0; i < number; i += 1) {
    text += '|- align="right"\n'
    text += `| "${rows[0]?.[i] ?? ''}"`
    for (let j = 2; j < 9; j += 1) {
      text += ` || ${rows[j]?.[i] ?? ''}`
    }
    text += '\n'
  }

  text += '|}'
  return text
}
