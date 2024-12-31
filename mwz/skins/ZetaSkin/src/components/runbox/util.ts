import { Md5 } from 'ts-md5'

import type { Job } from './types'

const queue: Promise<void>[] = [Promise.resolve()]

export function enqueue(f: Function, j: Job) {
  const task = queue.pop()
  if (task === undefined) return
  queue.push(task.then(() => new Promise((resolve) => { f(j, resolve) })))
}

export function wrap(el: Element, tag: string): Element {
  const wrapper = document.createElement(tag)
  el.parentNode?.insertBefore(wrapper, el)
  wrapper.appendChild(el)
  return wrapper
}

export function md5(obj: any) {
  return Md5.hashStr(JSON.stringify(obj))
}
