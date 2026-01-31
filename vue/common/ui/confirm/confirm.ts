// confirm.ts
import { reactive } from 'vue'

export type ConfirmColor = 'ghost' | 'default' | 'danger' | 'primary'

const defaultOptions = {
  okText: '확인',
  okColor: 'danger' as ConfirmColor,
  cancelText: '취소',
  closable: true,
}

export const confirmState = reactive({
  show: false,
  message: '',
  ...defaultOptions,
})

let resolver: ((v: boolean) => void) | null = null

function resolve(v: boolean) {
  confirmState.show = false
  resolver?.(v)
  resolver = null
}

export function showConfirm(message: string, options = {}) {
  Object.assign(confirmState, defaultOptions, options, {
    message,
    show: true,
  })
  return new Promise<boolean>(resolve => (resolver = resolve))
}

export function handleConfirmOk() {
  resolve(true)
}

export function handleConfirmCancel() {
  resolve(false)
}
