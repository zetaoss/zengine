export type DisambigNode = {
  description: string
  href: string
  id?: number
  new?: 1
  text: string
  title: string
}

export type Disambig = {
  id: number
  nodes: DisambigNode[]
  text: string
}
