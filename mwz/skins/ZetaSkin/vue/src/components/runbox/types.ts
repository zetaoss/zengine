// types.ts
export enum BoxType {
  Zero = '',
  Run = 'run',
  Notebook = 'notebook',
}

export enum JobType {
  Zero = '',
  Lang = 'lang',
  Front = 'front',
  Notebook = 'notebook',
}

export interface Box {
  type: BoxType
  jobId: string
  lang: string
  file: string
  title: string
  text: string
  isMain: boolean
  isAsciinema: boolean
  outResize: boolean
  el: Element
}

export interface Payload {
  lang: string
  files?: {
    name?: string
    body: string
  }[]
  main?: number
  sources?: string[]
}

export interface Output {
  output_type: string
  text?: string[]
  data?: Record<string, string[]>
  ename?: string
  evalue?: string
  traceback?: string[]
}

export interface Response {
  cpu: number
  mem: number
  time: number
  outs: unknown
}

export interface LangOut {
  logs: string[]
  images: string[]
}

export interface Job {
  id: string
  hash: string
  boxes: Box[]
  pageId: number
  main: number
  type: JobType
  phase: string | null
  payload: Payload | null
  langOuts: LangOut | null
  notebookOuts: Output[][]
  outResize: boolean
}
