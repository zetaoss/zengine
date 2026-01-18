// toast.ts
import { ref } from 'vue'

export interface ToastItem {
  id: number
  message: string
  timeout: number
}

export const toasts = ref<ToastItem[]>([])
let seed = 0

export function showToast(message: string, timeout = 2000) {
  const id = seed++
  toasts.value.push({ id, message, timeout })

  setTimeout(() => {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }, timeout)
}
