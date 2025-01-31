import { createApp, reactive } from 'vue';

import http from '@/utils/http';
import getRLCONF from '@/utils/rlconf';

import TheBox from './TheBox.vue';
import { type Box, BoxType, type Job, JobType, Step, } from './types';
import { enqueue, md5, wrap } from './util';

const boxes: Box[] = [];
const jobs: Job[] = reactive([]);
const pageId = getRLCONF().wgArticleId;

let delay = 1000;

const actions = {
  get: async (job: Job, resolve: () => void) => {
    try {
      const resp = await http.get(`/api/runbox/${job.hash}`);
      job.step = resp.step;
      switch (job.step) {
        case Step.Initial:
          enqueue(actions.post, job);
          break;
        case Step.Queued:
        case Step.Active:
          console.log(job.id, delay)
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
      }
    } catch (error) {
      console.error('Error in get:', error);
    } finally {
      resolve();
    }
  },
  post: async (job: Job, resolve: () => void) => {
    try {
      const data = {
        hash: job.hash,
        user_id: 0,
        page_id: job.pageId,
        type: job.type,
        payload: job.payload,
      }
      await http.post(`/api/runbox`, data);
      job.step = Step.Queued;
      enqueue(actions.get, job);
    } catch (error) {
      console.error('Error in post:', error);
    } finally {
      resolve();
    }
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
        payload: null,
        resp: null,
      });
    }
  })
}

function setJobs() {
  for (const job of jobs) {
    job.main = job.boxes.findIndex(b => b.isMain);
    if (job.main === -1) job.main = job.boxes.length - 1;
    const mainBox = job.boxes[job.main as number];
    const t = mainBox.type;
    if (t === BoxType.Run) {
      if (['javascript', 'html'].includes(mainBox.lang)) {
        job.type = JobType.Front;
      } else {
        job.type = JobType.Lang;
        job.payload = {
          lang: mainBox.lang,
          files: job.boxes.map(b => ({ name: b.file, body: b.text })),
          main: job.main,
        };
      }
    } else if (t === BoxType.Notebook) {
      job.type = JobType.Notebook
      job.payload = {
        lang: mainBox.lang,
        sources: job.boxes.map(b => b.text),
      }
    }
    job.hash = md5(job)
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
  for (const job of jobs) {
    if ([JobType.Notebook, JobType.Lang].includes(job.type)) {
      enqueue(actions.get, job);
    }
  }
}

export default function runbox() {
  createBoxes()
  createJobs()
  setJobs()
  runJobs()
}
