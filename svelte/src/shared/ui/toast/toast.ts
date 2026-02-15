import { writable } from 'svelte/store'

export interface ToastItem {
  id: number
  message: string
  timeout: number
}

export const toasts = writable<ToastItem[]>([])
let seed = 0

export function showToast(message: string, timeout = 2000) {
  const id = seed++
  toasts.update((items) => [...items, { id, message, timeout }])

  setTimeout(() => {
    toasts.update((items) => items.filter((t) => t.id !== id))
  }, timeout)
}
