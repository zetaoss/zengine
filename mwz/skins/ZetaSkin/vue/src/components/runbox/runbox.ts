import { createApp, reactive } from 'vue';
import http from '@/utils/http';
import getRLCONF from '@/utils/rlconf';
import TheBox from './TheBox.vue';
import { type Box, BoxType, type Job, JobType, Step } from './types';
import { enqueue, md5, wrap } from './util';

const jobs: Job[] = reactive([]);
const pageId = getRLCONF().wgArticleId;
let delay = 1000;

async function getJob(job: Job, resolve: () => void) {
  try {
    const { step, outs } = await http.get(`/api/runbox/${job.hash}`);
    job.step = step;

    if (step === Step.Initial) {
      return enqueue(postJob, job);
    }

    if (step === Step.Queued || step === Step.Active) {
      console.log(`Job ${job.id}: Retrying in ${delay}ms`);
      delay *= 1.1;
      setTimeout(() => enqueue(getJob, job), delay);
      return;
    }

    if (step === Step.Succeeded) {
      job[job.type === JobType.Lang ? 'logs' : 'outs'] = outs;
    }
  } catch (error) {
    console.error(`Error fetching job ${job.id}:`, error);
  } finally {
    resolve();
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
    });
    job.step = Step.Queued;
    enqueue(getJob, job);
  } catch (error) {
    console.error(`Error posting job ${job.id}:`, error);
  } finally {
    resolve();
  }
}

export default function runbox() {
  document.querySelectorAll('.mw-highlight').forEach((el, idx) => {
    const lang = el.className.match(/mw-highlight-lang-([a-z0-9]+)/)?.[1] || '';
    const run = el.getAttribute('run');
    const notebookAttr = el.getAttribute('notebook');
    const notebook = notebookAttr === '' ? 'noname' : notebookAttr;

    let boxType = BoxType.Zero;
    let jobId = `none-${idx}`;

    if (run != null) {
      boxType = BoxType.Run;
      jobId = `run-${run || `single-${idx}`}`;
    } else if (notebookAttr != null) {
      boxType = BoxType.Notebook;
      jobId = `notebook-${lang}-${notebook}`;
    }

    const box: Box = { type: boxType, jobId, lang, file: el.getAttribute('file') || '', title: el.getAttribute('title') || '', isMain: el.hasAttribute('main'), isAsciinema: el.hasAttribute('asciinema'), text: el.textContent || '', el };
    const job = jobs.find(j => j.id === jobId) || jobs[jobs.push({ id: jobId, type: JobType.Zero, hash: '', boxes: [], pageId, main: -1, step: Step.Initial, payload: null, logs: [], outs: [] }) - 1];
    job.boxes.push(box);
  });

  jobs.forEach(job => {
    const mainIdx = job.boxes.findIndex(b => b.isMain);
    job.main = mainIdx !== -1 ? mainIdx : job.boxes.length - 1;
    const mainBox = job.boxes[job.main];

    if (mainBox.type === BoxType.Run) {
      if (['javascript', 'html'].includes(mainBox.lang)) {
        job.type = JobType.Front;
      } else {
        job.type = JobType.Lang;
        const fileMap = new Map();
        job.boxes.forEach(b => {
          if (fileMap.has(b.file)) fileMap.get(b.file).body += b.text;
          else fileMap.set(b.file, { name: b.file || undefined, body: b.text });
        });
        job.payload = {
          lang: mainBox.lang,
          files: [...fileMap.values()],
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

    job.hash = md5(job.id, job.pageId, job.type, job.payload);

    job.boxes.forEach((box, seq) => {
      createApp({})
        .provide('job', job)
        .provide('seq', seq)
        .component('TheBox', TheBox)
        .mount(wrap(wrap(box.el, 'the-box'), 'div'));
    });

    if ([JobType.Notebook, JobType.Lang].includes(job.type)) {
      enqueue(getJob, job);
    }
  });
}
