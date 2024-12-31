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

export enum PhaseType {
  Init = 0,
  Queued = 1,
  Running = 2,
  Completed = 3,
  Error = 9,
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
  pageid: number
  main: number
  step: PhaseType
  reqRun?: ReqRun
  reqNotebook?: ReqNotebook
}
