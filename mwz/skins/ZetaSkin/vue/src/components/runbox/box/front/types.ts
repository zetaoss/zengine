export interface Word {
  typ: string
  text?: string
  words?: Word[]
  entries?: any[]
}

export interface Line {
  sev: string
  words: Word[]
}
