// runbox.helpers.ts
import type { Job } from './types'

let queue: Promise<void> = Promise.resolve()

type JobHandler = (job: Job) => Promise<void>

export function enqueue(handler: JobHandler, job: Job) {
  queue = queue
    .then(() => handler(job))
    .catch(e => {
      console.error('enqueue error', e)
    })
}

export function buildLangPayload(job: Job) {
  const mainBox = job.boxes[job.main] ?? job.boxes[0]
  const fileMap = new Map<string, { name?: string; body: string }>()

  if (!mainBox) {
    return {
      lang: '',
      files: [],
      main: job.main,
    }
  }

  job.boxes.forEach(b => {
    const key = b.file
    const existing = fileMap.get(key)

    if (existing) existing.body += b.text
    else fileMap.set(key, { name: key || undefined, body: b.text })
  })

  return {
    lang: mainBox.lang,
    files: [...fileMap.values()],
    main: job.main,
  }
}

export function buildNotebookPayload(job: Job) {
  const mainBox = job.boxes[job.main] ?? job.boxes[0]

  if (!mainBox) {
    return {
      lang: '',
      sources: [],
    }
  }

  return {
    lang: mainBox.lang,
    sources: job.boxes.map(b => b.text),
  }
}

export async function sha256(...args: unknown[]) {
  const msgBuffer = new TextEncoder().encode(JSON.stringify(args))
  const hashBuffer = await crypto.subtle.digest('SHA-256', msgBuffer)
  const hashArray = Array.from(new Uint8Array(hashBuffer))
  return hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
}
