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

export enum Step {
  Initial = 0,
  Queued = 1,
  Active = 2,
  Succeeded = 3,
  Failed = 9,
}

export interface Box {
  type: BoxType;
  jobId: string;
  lang: string;
  file: string;
  title: string;
  text: string;
  isMain: boolean;
  isAsciinema: boolean;
  el: Element;
}

export interface Payload {
  lang: string;
  files?: {
    name?: string;
    body: string;
  }[];
  main?: number;
  sources?: string[];
}

export interface Output {
  output_type: string;
  text?: string[];
  data?: Record<string, string[]>;
  ename?: string;
  evalue?: string;
  traceback?: string[];
}

export interface Response {
  cpu: number;
  mem: number;
  time: number;
  outs: unknown;
}

export interface Job {
  id: string;
  hash: string;
  boxes: Box[];
  pageId: number;
  main: number;
  type: JobType;
  step: Step;
  payload: Payload | null;
  logs: string[];
  outs: Output[][];
}

