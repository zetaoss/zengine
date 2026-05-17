// runbox.ts
import { mount } from 'svelte'
import { get, type Writable, writable } from 'svelte/store'

import getRLCONF from '$lib/utils/rlconf'
import httpy from '$shared/utils/httpy'

import BoxApex from './BoxApex.svelte'
import { buildLangPayload, buildNotebookPayload, enqueue, sha256 } from './runbox.helpers'
import { type Box, BoxType, type Job, JobType } from './types'

const pageId = getRLCONF().wgArticleId

let delay = 1000

interface JobStatus {
  phase: Job['phase'] | 'none' | 'Pending' | 'Running' | 'Succeeded' | 'error'
  outs: string | null
}

function parseOuts(raw: string | null): unknown {
  if (raw == null || raw.trim() === '') return null
  try {
    return JSON.parse(raw)
  } catch {
    return null
  }
}

function normalizeLangOuts(parsed: unknown): { logs: string[]; images: string[] } {
  const obj = parsed && typeof parsed === 'object' ? (parsed as Record<string, unknown>) : {}
  const logsRaw = obj.logs
  const imagesRaw = obj.images

  const logs = Array.isArray(logsRaw) ? logsRaw.map((v) => String(v)) : []
  const images = Array.isArray(imagesRaw) ? imagesRaw.map((v) => String(v)) : []
  return { logs, images }
}

function normalizeNotebookOuts(parsed: unknown): Job['notebookOuts'] {
  if (!Array.isArray(parsed)) {
    return []
  }

  return parsed.map((cell) => (Array.isArray(cell) ? cell : [])) as Job['notebookOuts']
}

type JobStore = Writable<Job>

const updateJob = (store: JobStore, updater: (job: Job) => void) => {
  store.update((job) => {
    updater(job)
    return job
  })
}

async function getJob(store: JobStore): Promise<void> {
  const job = get(store)
  const [data, err] = await httpy.get<JobStatus>(`/api/runbox/${job.hash}`)
  if (err) {
    console.error(err)
    updateJob(store, (j) => {
      j.phase = 'error'
    })
    return
  }

  const { phase, outs } = data
  const parsedOuts = parseOuts(outs)
  updateJob(store, (j) => {
    j.phase = phase
  })

  if (phase === 'none') {
    enqueue((j) => postJob(j), store)
    return
  }

  if (phase === 'Pending' || phase === 'Running') {
    console.log(`Job ${job.id}: ${phase}, refresh in ${Math.round(delay)}ms`)
    delay *= 1.1
    setTimeout(() => enqueue((j) => getJob(j), store), delay)
    return
  }

  if (phase === 'Failed') {
    updateJob(store, (j) => {
      if (j.type === JobType.Lang && j.langOuts == null) {
        j.langOuts = { logs: ['2Runbox job failed.'], images: [] }
      }
      if (j.type === JobType.Notebook && (!Array.isArray(j.notebookOuts) || j.notebookOuts.length === 0)) {
        j.notebookOuts = [[{ output_type: 'stream', text: ['Runbox job failed.'] }]]
      }
    })
    return
  }

  if (phase === 'Succeeded') {
    const jobType = get(store).type
    if (jobType === JobType.Lang) {
      updateJob(store, (j) => {
        j.langOuts = normalizeLangOuts(parsedOuts)
      })
    } else if (jobType === JobType.Notebook) {
      updateJob(store, (j) => {
        j.notebookOuts = normalizeNotebookOuts(parsedOuts)
      })
    }
  }
}

async function postJob(store: JobStore): Promise<void> {
  const job = get(store)
  const [, err] = await httpy.post('/api/runbox', {
    hash: job.hash,
    user_id: 0,
    page_id: job.pageId,
    type: job.type,
    payload: job.payload,
  })
  if (err) {
    console.error(err)
    updateJob(store, (j) => {
      j.phase = 'error'
    })
    return
  }

  updateJob(store, (j) => {
    j.phase = 'Pending'
  })
  enqueue((j) => getJob(j), store)
}

export async function rerunJob(store: JobStore): Promise<void> {
  const job = get(store)
  const [, err] = await httpy.post(`/api/runbox/${job.hash}/rerun`, {})
  if (err) {
    console.error(err)
    updateJob(store, (j) => {
      j.phase = 'error'
    })
    return
  }

  delay = 1000
  updateJob(store, (j) => {
    j.phase = 'Pending'
  })
  enqueue((j) => getJob(j), store)
}

export function mountRunbox() {
  const highlights = document.querySelectorAll<HTMLElement>('.mw-highlight')
  if (!highlights.length) return

  const jobsById = new Map<string, Job>()

  highlights.forEach((el, idx) => {
    const lang = el.className.match(/mw-highlight-lang-([a-z0-9]+)/)?.[1] || ''
    const run = el.getAttribute('run')
    const notebookAttr = el.getAttribute('notebook')
    const notebook = notebookAttr === '' ? 'noname' : notebookAttr

    let boxType = BoxType.Zero
    let jobId = `none-${idx}`

    if (run != null) {
      boxType = BoxType.Run
      jobId = `run-${run || `single-${idx}`}`
    } else if (notebookAttr != null) {
      boxType = BoxType.Notebook
      jobId = `notebook-${lang}-${notebook}`
    }

    const box: Box = {
      type: boxType,
      jobId,
      lang,
      file: el.getAttribute('file') || '',
      title: el.getAttribute('title') || '',
      isMain: el.hasAttribute('main'),
      isAsciinema: el.hasAttribute('asciinema'),
      outResize: el.hasAttribute('outresize'),
      text: el.textContent || '',
      el,
    }

    let job = jobsById.get(jobId)
    if (!job) {
      job = {
        id: jobId,
        type: JobType.Zero,
        hash: '',
        boxes: [],
        pageId,
        main: -1,
        phase: null,
        payload: null,
        langOuts: null,
        notebookOuts: [],
        outResize: box.outResize,
      }
      jobsById.set(jobId, job)
    }

    job.boxes.push(box)
  })

  const jobStores = [...jobsById.values()].map((job) => writable(job))

  jobStores.forEach(async (store) => {
    const job = get(store)
    if (!job.boxes.length) return

    const mainIdx = job.boxes.findIndex((b) => b.isMain)
    job.main = mainIdx !== -1 ? mainIdx : job.boxes.length - 1

    const mainBox = job.boxes[job.main]
    if (!mainBox) return

    if (mainBox.type === BoxType.Run) {
      if (['javascript', 'html', 'css'].includes(mainBox.lang)) {
        job.type = JobType.Front
      } else {
        job.type = JobType.Lang
        job.payload = buildLangPayload(job)
      }
    } else if (mainBox.type === BoxType.Notebook) {
      job.type = JobType.Notebook
      job.payload = buildNotebookPayload(job)
    } else {
      job.type = JobType.Zero
    }

    if (job.type === JobType.Lang || job.type === JobType.Notebook) {
      job.hash = await sha256(job.id, job.pageId, job.type, job.payload)
    }

    store.set(job)

    job.boxes.forEach((box, seq) => {
      mountBox(store, box.el, seq)
    })

    if (job.type === JobType.Lang || job.type === JobType.Notebook) {
      enqueue((j) => getJob(j), store)
    }
  })
}

function mountBox(store: JobStore, el: Element, seq: number) {
  const originalHTML = el.innerHTML
  el.innerHTML = ''

  mount(BoxApex, {
    target: el as HTMLElement,
    props: {
      job: store,
      seq,
      content: originalHTML,
    },
  })
}
