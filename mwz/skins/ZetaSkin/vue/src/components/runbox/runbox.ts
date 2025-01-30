import { createApp, reactive } from 'vue';

import http from '@/utils/http';
import getRLCONF from '@/utils/rlconf';

import TheBox from './TheBox.vue';
import {
  type Box,
  BoxType,
  type Job,
  JobType,
  Step,
} from './types';
import { enqueue, md5, wrap } from './util';

const boxes: Box[] = [];
const jobs: Job[] = reactive([]);
const pageId = getRLCONF().wgArticleId;

let delay = 1000;

const actions = {
  get: async (job: Job, resolve: () => void) => {
    console.log('get', job);
    const resp = await http.get(`/api/runbox/${pageId}/${md5(job)}`);
    job.step = resp.step
    switch (job.step) {
      case Step.Initial:
        enqueue(actions.post, job);
        break;
      case Step.Active:
        delay *= 1.1;
        setTimeout(() => enqueue(actions.get, job), delay);
        break;
      case Step.Succeeded:
        job.resp = resp;
        job.boxes.forEach((b, i) => {
          if (b.jobId == job.id && i == job.main) {
            console.log('pick b', b)
          }
        });
        break;
      case Step.Failed:
        break;
      default:
    }
    resolve();
  },
  post: async (job: Job, resolve: () => void) => {
    console.log('===> post job:', job);
    switch (job.type) {
      case JobType.Run:
        try {
          const resp = await http.post(`/api/runbox/${pageId}/${md5(job)}`, job.reqRun)
          job.step = resp.step
          if (job.step < Step.Succeeded) enqueue(actions.get, job)
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
    let boxType = BoxType.Zero;
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
        type: JobType.Zero,
        hash: '',
        boxes: [b],
        pageId,
        main: -1,
        step: Step.Initial,
        resp: null,
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
    for (const [seq, box] of job.boxes.entries()) {
      const app = createApp({});
      app.provide('job', job);
      app.provide('seq', seq);
      app.provide('box', box);
      app.component('TheBox', TheBox);
      app.mount(wrap(wrap(box.el, 'the-box'), 'div'));
    }
  }
}


function runJobs() {
  const runnableJobTypes = [JobType.Notebook, JobType.Run];
  for (const job of jobs) {
    if (runnableJobTypes.includes(job.type)) {
      enqueue(actions.get, job);
    }
  }
}

export default function runbox() {
  createBoxes()
  createJobs()
  setJobs()
  runJobs()

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
}
