import { writable } from 'svelte/store'

export interface ToastItem {
  id: number
  message: string
  timeout: number
  placement: 'top-right' | 'center'
  action?: {
    label: string
    href: string
  }
}

export const toasts = writable<ToastItem[]>([])
let seed = 0

export interface ToastOptions {
  timeout?: number
  placement?: ToastItem['placement']
  action?: ToastItem['action']
}

export function dismissToast(id: number) {
  toasts.update((items) => items.filter((t) => t.id !== id))
}

export function showToast(message: string, timeoutOrOptions: number | ToastOptions = 2000) {
  const id = seed++
  const options = typeof timeoutOrOptions === 'number' ? { timeout: timeoutOrOptions } : timeoutOrOptions
  const timeout = options.timeout ?? 2000

  toasts.update((items) => [
    ...items,
    {
      id,
      message,
      timeout,
      placement: options.placement ?? 'top-right',
      action: options.action,
    },
  ])

  setTimeout(() => {
    dismissToast(id)
  }, timeout)
}
