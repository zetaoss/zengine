// useConfirm.ts
import { reactive } from 'vue'

export type ConfirmColor = 'ghost' | 'default' | 'danger' | 'primary'

const defaultOptions = {
  okText: '확인',
  okColor: 'danger' as ConfirmColor,
  cancelText: '취소',
  closable: true,
}

const state = reactive({
  show: false,
  message: '',
  ...defaultOptions,
})

let resolver: ((v: boolean) => void) | null = null

export function useConfirm() {
  return (message: string, options = {}) => {
    Object.assign(state, defaultOptions, options, {
      message,
      show: true,
    })
    return new Promise<boolean>((resolve) => (resolver = resolve))
  }
}

function resolve(v: boolean) {
  state.show = false
  resolver?.(v)
  resolver = null
}

export function useConfirmController() {
  return {
    state,
    handleOk: () => resolve(true),
    handleCancel: () => resolve(false),
  }
}
