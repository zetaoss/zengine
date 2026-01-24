// @/types/toc.ts

export interface Section {
  toclevel: number
  level: string
  line: string
  number: string
  index: string
  fromtitle: string
  byteoffset: number
  anchor: string
  'array-sections': Section[]
  'is-top-level-section': boolean
  'is-parent-section': boolean
}
