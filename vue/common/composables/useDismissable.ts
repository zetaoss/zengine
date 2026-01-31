// @common/composables/useDismissable.ts
import type { Ref } from 'vue'
import { onBeforeUnmount, onMounted, unref } from 'vue'

interface Options {
  enabled: Ref<boolean>
  onDismiss: () => void
}

export function useDismissable(
  target: Ref<HTMLElement | null>,
  { enabled, onDismiss }: Options,
) {
  function onMouseDown(e: MouseEvent) {
    if (!unref(enabled)) return

    const el = unref(target)
    if (!el) return
    if (el.contains(e.target as Node)) return

    onDismiss()
  }

  function onKeyDown(e: KeyboardEvent) {
    if (!unref(enabled)) return
    if (e.key === 'Escape') onDismiss()
  }

  onMounted(() => {
    document.addEventListener('mousedown', onMouseDown)
    document.addEventListener('keydown', onKeyDown)
  })

  onBeforeUnmount(() => {
    document.removeEventListener('mousedown', onMouseDown)
    document.removeEventListener('keydown', onKeyDown)
  })
}
