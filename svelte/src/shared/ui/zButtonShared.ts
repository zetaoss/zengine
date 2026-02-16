export type ZButtonColor = 'default' | 'danger' | 'ghost' | 'primary'
export type ZButtonSize = 'small' | 'medium'

const sizeClasses: Record<ZButtonSize, string> = {
  medium: 'p-2 text-sm',
  small: 'p-1 text-xs',
}

const baseClass =
  'text-[var(--z-text)] inline-flex cursor-pointer items-center justify-center rounded transition ring-1 hover:no-underline leading-none'

const colorClasses: Record<ZButtonColor, string> = {
  default: 'bg-[var(--z-btn-bg)] ring-[var(--z-btn-hover)] hover:bg-[var(--z-btn-hover)]',
  danger: 'bg-[var(--z-danger-bg)] ring-[var(--z-danger-hover)] hover:bg-[var(--z-danger-hover)]',
  ghost: 'bg-transparent ring-transparent hover:bg-[var(--z-btn-hover)]',
  primary: 'bg-[var(--z-primary-bg)] ring-[var(--z-primary-hover)] hover:bg-[var(--z-primary-hover)]',
}

const disabledClass = 'opacity-50 brightness-[.9] cursor-not-allowed pointer-events-none text-[var(--z-btn-text-disabled)]'

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
