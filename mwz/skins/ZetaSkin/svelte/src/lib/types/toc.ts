export type Section = {
  anchor: string
  byteoffset: number
  index: string
  fromtitle: string
  level: string
  line: string
  linkAnchor?: string
  number: string
  toclevel: number
  'array-sections': Section[]
  'is-parent-section': boolean
  'is-top-level-section': boolean
}

export type DataToc = {
  'array-sections': Section[]
  'number-section-count'?: number
}
