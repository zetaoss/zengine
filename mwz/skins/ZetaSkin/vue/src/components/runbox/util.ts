import { Md5 } from 'ts-md5'

import type { Job } from './types'

const queue: Promise<void>[] = [Promise.resolve()]

type JobHandler = (job: Job, done: () => void) => void;

export function enqueue(f: JobHandler, j: Job) {
  if (queue.length === 0) return;

  const task = queue.pop();
  if (!task) return;

  queue.push(
    task.then(
      () => new Promise<void>((resolve) => f(j, resolve))
    )
  );
}

export function wrap(el: Element, tag: string): Element {
  const wrapper = document.createElement(tag)
  el.parentNode?.insertBefore(wrapper, el)
  wrapper.appendChild(el)
  return wrapper
}

export function md5(...args: unknown[]): string {
  return Md5.hashStr(JSON.stringify(args));
}
