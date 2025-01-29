export enum BoxType {
  None = '',
  Run = 'run',
  Notebook = 'notebook',
}

export enum JobType {
  None = '',
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
    text: string
  }[]
}

export interface ReqNotebook {
  lang: string
  cellTexts: string[][]
}

export interface Job {
  id: string
  type: JobType
  boxes: Box[]
  pageId: number
  main: number
  state: StateType
  reqRun?: ReqRun
  reqNotebook?: ReqNotebook
}
