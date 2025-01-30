export enum BoxType {
  Zero = '',
  Run = 'run',
  Notebook = 'notebook',
}

export enum JobType {
  Zero = '',
  Run = 'run',
  Front = 'front',
  Notebook = 'notebook',
}

export enum StateType {
  Initial = 0,
  Queued = 1,
  Active = 2,
  Succeeded = 3,
  Failed = 9,
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
  el: Element
}

export interface ReqRun {
  lang: string
  main: number
  files: {
    name: string
    body: string
  }[]
}

export interface ReqNotebook {
  lang: string
  cellTexts: string[][]
}

export interface Resp {
  cpu: number
  mem: number
  time: number
  logs: string[]
}

export interface Job {
  id: string
  type: JobType
  hash: string
  boxes: Box[]
  pageId: number
  main: number
  state: StateType
  reqRun?: ReqRun
  reqNotebook?: ReqNotebook
  resp: Resp | null
}
