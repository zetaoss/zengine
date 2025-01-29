import { createApp } from 'vue';

import http from '@/utils/http';
import getRLCONF from '@/utils/rlconf';

import TheBox from './TheBox.vue';
import {
  type Box,
  BoxType,
  type Job,
  JobType,
  StateType,
} from './types';
import { enqueue, md5, wrap } from './util';

const boxes: Box[] = [];
const jobs: Job[] = [];
const pageId = getRLCONF().wgArticleId;

let delay = 1000;

const actions = {
  get: async (job: Job, resolve: () => void) => {
    console.log('get', job);
    const resp = await http.get(`/api/runbox/${pageId}/${md5(job)}`);
    console.log('resp', resp);
    job.state = resp.state
    switch (job.state) {
      case StateType.Initial:
        enqueue(actions.post, job);
        break;
      case StateType.Active:
        delay *= 1.1;
        console.log('delay=', delay);
        setTimeout(() => enqueue(actions.get, job), delay);
        break
      case StateType.Failed:
        // job.message = resp.message;
        break
      default:
    }
    resolve();
  },
  post: async (job: Job, resolve: () => void) => {
    console.log('=> post', job);
    switch (job.type) {
      case JobType.Run:
        try {
          const resp = await http.post(`/api/runbox/${pageId}/${md5(job)}`, job.reqRun)
          job.state = resp.state
          if (job.state < StateType.Succeeded) enqueue(actions.get, job)
        } catch (e) {
          console.error('err', e)
        }
    }
    resolve()
  },
}

function createBoxes() {
  const els = document.getElementsByClassName('mw-highlight')
  Array.from(els).forEach((el, idx) => {
    // lang
    const cls = el.getAttribute('class') ?? '';
    const langMatched = cls.match(/mw-highlight-lang-([a-z0-9]+)/);
    const lang = (langMatched == null) ? '' : langMatched[1];
    // type & group
    const run = el.getAttribute('run');
    const notebook = el.getAttribute('notebook');
    let boxType = BoxType.None;
    let jobId = `none-${idx}`;
    if (run != null) {
      boxType = BoxType.Run;
      jobId = (run === '') ? `run-single-${idx}` : `run-${run}`;
    } else if (notebook != null) {
      boxType = BoxType.Notebook;
      jobId = (notebook === '') ? `notebook-single-${idx}` : `notebook-${lang}-${notebook}`;
    }
    boxes.push({
      type: boxType,
      jobId,
      lang,
      file: el.getAttribute('file') ?? '',
      title: el.getAttribute('title') ?? '',
      isMain: !(el.getAttribute('main') === null),
      isAsciinema: !(el.getAttribute('asciinema') === null),
      text: el.textContent!,
      el,
    });
  })
}

function createJobs() {
  boxes.forEach((b) => {
    if (jobs.length > 0 && jobs[jobs.length - 1].id === b.jobId) {
      jobs[jobs.length - 1].boxes.push(b);
    } else {
      jobs.push({
        id: b.jobId,
        type: JobType.None,
        hash: "",
        boxes: [b],
        pageId,
        main: -1,
        state: StateType.Initial,
      });
    }
  })
}

function setJobs() {
  for (const job of jobs) {
    job.main = job.boxes.findIndex((b) => b.isMain);
    if (job.main === -1) job.main = job.boxes.length - 1;
    // set jobs
    const mainBox = job.boxes[job.main as number];
    const t = mainBox.type;
    if (t === BoxType.Run) {
      if (['javascript', 'html'].includes(mainBox.lang)) {
        job.type = JobType.Front;
      } else {
        job.type = JobType.Run;
        const req = {
          lang: mainBox.lang,
          main: job.main,
          files: job.boxes.map((b) => ({ name: b.file, body: b.text })),
        }
        job.reqRun = req;
        job.hash = md5(req);
      }
    } else if (t === BoxType.Notebook) {
      job.type = JobType.Notebook
      job.reqNotebook = {
        lang: mainBox.lang,
        cellTexts: job.boxes.map((b) => [b.text]),
      }
    }
    // render box
    for (const [i, b] of job.boxes.entries()) {
      const app = createApp({})
      app.provide('job', job);
      app.provide('seq', i);
      app.component('TheBox', TheBox);
      app.mount(wrap(wrap(b.el, 'the-box'), 'div'));
    }
  }
}


function runJobs() {
  const relevantJobTypes = [JobType.Notebook, JobType.Run];
  for (const job of jobs) {
    if (relevantJobTypes.includes(job.type)) {
      enqueue(actions.get, job);
    }
  }
}

export default function runbox() {
  createBoxes()
  createJobs()
  setJobs()
  runJobs()
  console.log('boxes', boxes)
  console.log('jobs', jobs)

  // console.log(boxes)
  // setJobValues()
  //
  // set jobs & boxes
  // refJobs.forEach((refJob, index) => {
  //   const job = refJob.value
  //   boxes.filter((b) => job.boxIds.includes(b.id)).forEach((b) => {
  //     b.spec.jobIndex = index
  //     b.refJob = refJob
  //     if (b.spec.isMain || job.lang === '') job.lang = b.spec.lang
  //     job.texts.push(b.spec.text)
  //   })
  //   job.api = getAPIType(job)
  //   job.hash = Md5.hashStr(job.api + job.lang + job.texts.join())
  //   enqueue(get, job)
  // })
  // // app mount
  // boxes.forEach((b) => {
  //   const { app } = b
  //   app.provide('box', b.spec)
  //   app.provide('job', b.refJob)
  //   app.mount(b.root)
  // })
}
