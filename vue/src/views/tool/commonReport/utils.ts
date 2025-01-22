import { useDateFormat } from '@vueuse/core'

import type { Row } from './types'

export function getRatio(row: Row, idx: number) {
  return row.items[idx].total / row.total
}

export function getScore(row: Row): number {
  const nums = row.items.map(x => x.total)
  const total = nums.reduce((acc, cur) => acc + cur, 0)
  if (nums.length < 2 || total == 0) return -1
  const p = nums.map(x => x / total)
  const p0 = p[0]
  const p1 = p[1]
  const gap = p0 - p1
  const sizeWeight = Math.log10(total)
  const weightedDominance = (gap * 0.8 + p0 * 0.2) * sizeWeight
  const maxWeight = Math.log10(1000000)
  return Math.floor(Math.max(0, Math.min(weightedDominance / maxWeight, 1)) * 4)
}

export function getWikitextTable(table: HTMLTableElement, id: number, url: string, createdAt: string) {
  // get rows
  let number = 0
  const rows = [] as string[][]
  const trs = table.getElementsByTagName('tr')
  for (let i = 0; i < trs.length; i++) {
    const tds = trs[i].getElementsByTagName('td')
    const row = [] as string[]
    for (let j = 0; j < tds.length; j++) {
      const td = tds[j] as HTMLElement
      row.push(td.innerText)
    }
    rows.push(row)
    number = row.length
  }
  // create text
  const dt = useDateFormat(createdAt, 'YYYY-MM-DD').value
  let text = `{| class='wikitable'
|+ [${url} 통용 보고서 #${id} ( ${dt} )]
|-
! 표기 !! 비율 !! 계 !! D블로그 !! N블로그 !! N책 !! N뉴스 !! 구글
`
  for (let i = 0; i < number; i++) {
    text += '|- align="right"\n'
    text += `| "${rows[0][i]}"`
    for (let j = 2; j < 9; j++) {
      text += ` || ${rows[j][i]}`
    }
    text += '\n'
  }
  text += '|}'
  return text
}
