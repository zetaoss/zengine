export interface Word {
  type: string
  text?: string
  list?: Word[]
  dict?: Record<string, Word>
}

export interface Line {
  sev: string
  words: Word[]
}

export interface Line0 {
  sev: string
  args: unknown[]
}
