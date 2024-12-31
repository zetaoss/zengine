import { useDateFormat } from '@vueuse/core'

import type { Row, Score } from './types'

export function getRatio(row: Row, idx: number) {
  return row.items[idx].total / row.total
}

export function getScore(row: Row): Score {
  const { items } = row
  const top2 = items[0].total + items[1].total
  if (top2 === 0) {
    return { star: 0, grade: '—' }
  }
  const quorum = items[0].total / top2
  const sampleSizeIndex = Math.log10(Math.log10(top2) + 1)
  const num = quorum ** 0.5 * sampleSizeIndex ** 1.5 * 5.5
  if (num < 0.1) return { star: 0, grade: 'D' }
  if (num < 1) return { star: 1, grade: 'C' }
  if (num < 2) return { star: 2, grade: 'B' }
  if (num < 3) return { star: 3, grade: 'A' }
  return { star: 4, grade: 'S' }
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
