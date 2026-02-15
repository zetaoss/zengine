import { onDestroy } from 'svelte'

type MaybeElement = HTMLElement | null | undefined

interface UseDismissableOptions {
  enabled?: () => boolean
  onDismiss: (event: Event) => void
  ignore?: () => MaybeElement[]
}

function isNode(value: unknown): value is Node {
  return value instanceof Node
}

export function useDismissable(getRoot: () => MaybeElement, options: UseDismissableOptions): void {
  if (typeof document === 'undefined') return

  const isEnabled = () => options.enabled?.() ?? true

  const isIgnoredTarget = (target: Node) => {
    const ignore = options.ignore?.() ?? []
    return ignore.some((el) => !!el && el.contains(target))
  }

  const onPointerDown = (event: Event) => {
    if (!isEnabled()) return

    const root = getRoot()
    const target = event.target
    if (!root || !isNode(target)) return
    if (root.contains(target) || isIgnoredTarget(target)) return

    options.onDismiss(event)
  }

  const onKeyDown = (event: Event) => {
    if (!isEnabled()) return

    const keyboard = event as KeyboardEvent
    if (keyboard.key !== 'Escape') return

    options.onDismiss(event)
  }

  document.addEventListener('pointerdown', onPointerDown, true)
  document.addEventListener('keydown', onKeyDown)

  onDestroy(() => {
    document.removeEventListener('pointerdown', onPointerDown, true)
    document.removeEventListener('keydown', onKeyDown)
  })
}
