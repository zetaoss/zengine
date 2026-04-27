export type ZButtonColor = 'default' | 'danger' | 'ghost' | 'primary'
export type ZButtonSize = 'small' | 'medium'

const sizeClasses: Record<ZButtonSize, string> = {
  medium: 'p-2 text-sm',
  small: 'p-1 text-xs',
}

const baseClass =
  'text-(--z-text) inline-flex cursor-pointer items-center justify-center rounded transition ring-1 hover:no-underline hover:shadow-[inset_0_0_0_9999px_#8884] leading-none'

const colorClasses: Record<ZButtonColor, string> = {
  default: 'bg-(--z-btn-bg) ring-(--z-btn-hover)',
  danger: 'bg-(--z-danger-bg) ring-(--z-danger-hover)',
  ghost: 'bg-transparent ring-transparent',
  primary: 'bg-(--z-primary-bg) ring-(--z-primary-hover)',
}

const disabledClass = 'opacity-50 brightness-[.9] cursor-not-allowed pointer-events-none text-(--z-btn-text-disabled)'

export function getZButtonClasses(opts: { size: ZButtonSize; color: ZButtonColor; isDisabled: boolean; className?: string }): string {
  return [baseClass, sizeClasses[opts.size], colorClasses[opts.color], opts.isDisabled ? disabledClass : '', opts.className ?? '']
    .filter(Boolean)
    .join(' ')
}

export function handleZButtonClick(opts: {
  event: MouseEvent
  isDisabled: boolean
  onclick?: (event: MouseEvent) => void
  cooldown?: number
  startCooldown?: (ms: number) => void
}): void {
  const { event, isDisabled, onclick, cooldown = 0, startCooldown } = opts

  if (isDisabled) {
    event.preventDefault()
    event.stopImmediatePropagation()
    return
  }

  onclick?.(event)

  if (cooldown > 0 && startCooldown) {
    startCooldown(cooldown)
  }
}
