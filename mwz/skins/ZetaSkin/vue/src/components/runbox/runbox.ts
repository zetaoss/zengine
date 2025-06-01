import { createApp, reactive, h } from 'vue'
import http from '@/utils/http'
import getRLCONF from '@/utils/rlconf'
import BoxVue from './Box.vue'
import { type Box, BoxType, type Job, JobType } from './types'

const jobs: Job[] = reactive([])
const pageId = getRLCONF().wgArticleId
let delay = 1000

const queue: Promise<void>[] = [Promise.resolve()]

type JobHandler = (job: Job, done: () => void) => void;

async function getJob(job: Job, resolve: () => void) {
  try {
    const { phase, outs } = await http.get(`/api/runbox/${job.hash}`)
    job.phase = phase

    if (phase === 'none') {
      return enqueue(postJob, job)
    }

    if (phase === 'pending' || phase === 'running') {
      console.log(`Job ${job.id}: Retrying in ${delay}ms`)
      delay *= 1.1
      setTimeout(() => enqueue(getJob, job), delay)
      return
    }

    if (phase === 'succeeded') {
      if (job.type === JobType.Lang) {
        console.log(`[${job.id}] phase=succeeded → setting logs`, outs);
        job.logs.splice(0, job.logs.length, ...outs);
      } else {
        job.outs = outs;
      }
    }
  } catch (error) {
    console.error(`Error fetching job ${job.id}:`, error)
  } finally {
    resolve()
  }
}

async function postJob(job: Job, resolve: () => void) {
  try {
    await http.post(`/api/runbox`, {
      hash: job.hash,
      user_id: 0,
      page_id: job.pageId,
      type: job.type,
      payload: job.payload,
    })
    job.phase = 'pending'
    enqueue(getJob, job)
  } catch (error) {
    console.error(`Error posting job ${job.id}:`, error)
  } finally {
    resolve()
  }
}

export function runbox() {
  document.querySelectorAll<HTMLElement>('.mw-highlight').forEach((el, idx) => {
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
      text: el.textContent || '',
      el
    }

    const job = jobs.find(j => j.id === jobId)
      || jobs[jobs.push({
        id: jobId,
        type: JobType.Zero,
        hash: '',
        boxes: [],
        pageId,
        main: -1,
        phase: null,
        payload: null,
        logs: [],
        outs: [],
      }) - 1]

    job.boxes.push(box)
  })

  jobs.forEach(async job => {
    const mainIdx = job.boxes.findIndex(b => b.isMain)
    job.main = mainIdx !== -1 ? mainIdx : job.boxes.length - 1
    const mainBox = job.boxes[job.main]

    if (mainBox.type === BoxType.Run) {
      if (['javascript', 'html'].includes(mainBox.lang)) {
        job.type = JobType.Front
      } else {
        job.type = JobType.Lang
        const fileMap = new Map()
        job.boxes.forEach(b => {
          if (fileMap.has(b.file)) {
            fileMap.get(b.file).body += b.text
          } else {
            fileMap.set(b.file, { name: b.file || undefined, body: b.text })
          }
        })
        job.payload = {
          lang: mainBox.lang,
          files: [...fileMap.values()],
          main: job.main,
        }
      }
    } else if (mainBox.type === BoxType.Notebook) {
      job.type = JobType.Notebook
      job.payload = {
        lang: mainBox.lang,
        sources: job.boxes.map(b => b.text),
      }
    }

    job.hash = await sha256(job.id, job.pageId, job.type, job.payload)

    job.boxes.forEach((box, seq) => {
      const el = box.el
      const originalHTML = el.innerHTML
      el.innerHTML = ''
      createApp({
        render() {
          return h(BoxVue, null, {
            default: () => h('div', { innerHTML: originalHTML })
          })
        }
      })
        .provide('job', job)
        .provide('seq', seq)
        .mount(el)
    })

    if ([JobType.Notebook, JobType.Lang].includes(job.type)) {
      enqueue(getJob, job)
    }
  })
}

function enqueue(f: JobHandler, j: Job) {
  const prev = queue.length > 0 ? queue[queue.length - 1] : Promise.resolve();
  queue.push(
    prev.then(() => new Promise<void>((resolve) => f(j, resolve)))
  );
}

async function sha256(...args: unknown[]) {
  const msgBuffer = new TextEncoder().encode(JSON.stringify(args));
  const hashBuffer = await crypto.subtle.digest('SHA-256', msgBuffer);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
  return hashHex;
}
