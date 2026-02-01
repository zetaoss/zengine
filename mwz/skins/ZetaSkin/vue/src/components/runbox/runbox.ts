// runbox.ts
import httpy from '@common/utils/httpy'
import { createApp, h, reactive } from 'vue'

import getRLCONF from '@/utils/rlconf'

import BoxApex from './BoxApex.vue'
import {
  buildLangPayload,
  buildNotebookPayload,
  enqueue,
  sha256,
} from './runbox.helpers'
import { type Box, BoxType, type Job, JobType } from './types'

const jobs: Job[] = reactive([])
const pageId = getRLCONF().wgArticleId

let delay = 1000

interface JobStatus {
  phase: Job['phase'] | 'none' | 'pending' | 'running' | 'succeeded' | 'error'
  outs: unknown
}

async function getJob(job: Job): Promise<void> {
  const [data, err] = await httpy.get<JobStatus>(`/api/runbox/${job.hash}`)
  if (err) {
    console.error(err)
    job.phase = 'error'
    return
  }

  const { phase, outs } = data
  job.phase = phase
  if (phase === 'none') {
    enqueue(postJob, job)
    return
  }

  if (phase === 'pending' || phase === 'running') {
    console.log(`Job ${job.id}: Retrying in ${delay}ms`)
    delay *= 1.1
    setTimeout(() => enqueue(getJob, job), delay)
    return
  }

  if (phase === 'succeeded') {
    if (job.type === JobType.Lang) {
      job.langOuts = outs as typeof job.langOuts
    } else if (job.type === JobType.Notebook) {
      job.notebookOuts = outs as typeof job.notebookOuts
    }
  }
}

async function postJob(job: Job): Promise<void> {
  const [, err] = await httpy.post('/api/runbox', {
    hash: job.hash,
    user_id: 0,
    page_id: job.pageId,
    type: job.type,
    payload: job.payload,
  })
  if (err) {
    console.error(err)
    job.phase = 'error'
    return
  }

  job.phase = 'pending'
  enqueue(getJob, job)
}

export function mountRunbox() {
  const highlights = document.querySelectorAll<HTMLElement>('.mw-highlight')
  if (!highlights.length) return

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

    const job =
      jobs.find(j => j.id === jobId) ??
      (() => {
        const newJob: Job = {
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
        jobs.push(newJob)
        return newJob
      })()

    job.boxes.push(box)
  })

  jobs.forEach(async job => {
    if (!job.boxes.length) return

    const mainIdx = job.boxes.findIndex(b => b.isMain)
    job.main = mainIdx !== -1 ? mainIdx : job.boxes.length - 1
    const mainBox = job.boxes[job.main]
    if (!mainBox) return

    if (mainBox.type === BoxType.Run) {
      if (['javascript', 'html'].includes(mainBox.lang)) {
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

    job.boxes.forEach((box, seq) => {
      mountBox(job, box.el, seq)
    })

    if (job.type === JobType.Lang || job.type === JobType.Notebook) {
      enqueue(getJob, job)
    }
  })
}

function mountBox(job: Job, el: Element, seq: number) {
  const originalHTML = el.innerHTML
  el.innerHTML = ''
  createApp({
    setup() {
      return () =>
        h(
          BoxApex,
          { job, seq },
          { default: () => h('div', { innerHTML: originalHTML }) },
        )
    },
  }).mount(el)
}
