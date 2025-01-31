import { createApp, reactive } from 'vue';
import http from '@/utils/http';
import getRLCONF from '@/utils/rlconf';
import TheBox from './TheBox.vue';
import { type Box, BoxType, type Job, JobType, Step } from './types';
import { enqueue, md5, wrap } from './util';

const jobs: Job[] = reactive([]);
const pageId = getRLCONF().wgArticleId;
let delay = 1000;

const actions = {
  async get(job: Job, resolve: () => void) {
    try {
      const response = await http.get(`/api/runbox/${job.hash}`);
      job.step = response.step;

      switch (job.step) {
        case Step.Initial:
          enqueue(actions.post, job);
          break;
        case Step.Queued:
        case Step.Active:
          console.log(`Job ${job.id}: Retrying in ${delay}ms`);
          delay *= 1.1;
          setTimeout(() => enqueue(actions.get, job), delay);
          break;
        case Step.Succeeded:
          job.response = response;
          job.boxes.forEach((b, i) => {
            if (b.jobId == job.id && i == job.main) {
              console.log('Main box selected:', b);
            }
          });
          break;
      }
    } catch (error) {
      console.error(`Error fetching job ${job.id}:`, error);
    } finally {
      resolve();
    }
  },

  async post(job: Job, resolve: () => void) {
    try {
      await http.post(`/api/runbox`, {
        hash: job.hash,
        user_id: 0,
        page_id: job.pageId,
        type: job.type,
        payload: job.payload,
      });
      job.step = Step.Queued;
      enqueue(actions.get, job);
    } catch (error) {
      console.error(`Error posting job ${job.id}:`, error);
    } finally {
      resolve();
    }
  },
}

function createJobs() {
  const boxes = createBoxes();
  boxes.forEach(b => {
    const lastJob = jobs[jobs.length - 1];
    if (lastJob && lastJob.id === b.jobId) {
      lastJob.boxes.push(b);
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
        response: null,
      });
    }
  })
}

function createBoxes(): Box[] {
  const boxes: Box[] = [];
  document.querySelectorAll('.mw-highlight').forEach((el, idx) => {
    const cls = el.getAttribute('class') ?? '';
    const langMatch = cls.match(/mw-highlight-lang-([a-z0-9]+)/);
    const lang = langMatch ? langMatch[1] : '';

    const run = el.getAttribute('run');
    const notebook = el.getAttribute('notebook');
    let boxType = BoxType.Zero;
    let jobId = `none-${idx}`;

    if (run != null) {
      boxType = BoxType.Run;
      jobId = run ? `run-${run}` : `run-single-${idx}`;
    } else if (notebook != null) {
      boxType = BoxType.Notebook;
      jobId = notebook ? `notebook-${lang}-${notebook}` : `notebook-single-${idx}`;
    }

    boxes.push({
      type: boxType,
      jobId,
      lang,
      file: el.getAttribute('file') ?? '',
      title: el.getAttribute('title') ?? '',
      isMain: el.hasAttribute('main'),
      isAsciinema: el.hasAttribute('asciinema'),
      text: el.textContent ?? '',
      el,
    });
  });
  return boxes;
}

function setJobs() {
  jobs.forEach(job => {
    job.main = job.boxes.findIndex(b => b.isMain);
    if (job.main === -1) job.main = job.boxes.length - 1;

    const mainBox = job.boxes[job.main];
    if (mainBox.type === BoxType.Run) {
      job.type = ['javascript', 'html'].includes(mainBox.lang) ? JobType.Front : JobType.Lang;
      if (job.type === JobType.Lang) {
        job.payload = {
          lang: mainBox.lang,
          files: job.boxes.map(b => ({ name: b.file, body: b.text })),
          main: job.main,
        };
      }
    } else if (mainBox.type === BoxType.Notebook) {
      job.type = JobType.Notebook
      job.payload = {
        lang: mainBox.lang,
        sources: job.boxes.map(b => b.text),
      }
    }
    job.hash = md5(job);
    renderJob(job);
  });
}

function renderJob(job: Job) {
  job.boxes.forEach((box, seq) => {
    const app = createApp({});
    app.provide('job', job);
    app.provide('seq', seq);
    app.provide('box', box);
    app.component('TheBox', TheBox);
    app.mount(wrap(wrap(box.el, 'the-box'), 'div'));
  });
}

function runJobs() {
  jobs.forEach((job) => {
    if ([JobType.Notebook, JobType.Lang].includes(job.type)) {
      enqueue(actions.get, job);
    }
  });
}

export default function runbox() {
  createJobs();
  setJobs();
  runJobs();
}
