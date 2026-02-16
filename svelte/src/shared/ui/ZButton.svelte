<script lang="ts">
  type Color = 'default' | 'danger' | 'ghost' | 'primary'

  export let as: string = 'button'
  export let disabled = false
  export let color: Color = 'default'
  export let size: 'small' | 'medium' = 'medium'
  export let cooldown = 500
  export let onclick: ((event: MouseEvent) => void) | undefined = undefined

  let isCooling = false

  const sizeClasses: Record<string, string> = {
    medium: 'p-2 text-sm',
    small: 'p-1 text-xs',
  }

  const base =
    'text-[var(--z-text)] inline-flex cursor-pointer items-center justify-center rounded transition ring-1 hover:no-underline leading-none'

  const colorClasses: Record<Color, string> = {
    default: 'bg-[var(--z-btn-bg)] ring-[var(--z-btn-hover)] hover:bg-[var(--z-btn-hover)]',
    danger: 'bg-[var(--z-danger-bg)] ring-[var(--z-danger-hover)] hover:bg-[var(--z-danger-hover)]',
    ghost: 'bg-transparent ring-transparent hover:bg-[var(--z-btn-hover)]',
    primary: 'bg-[var(--z-primary-bg)] ring-[var(--z-primary-hover)] hover:bg-[var(--z-primary-hover)]',
  }

  const disabledClasses = 'opacity-50 brightness-[.9] cursor-not-allowed pointer-events-none text-[var(--z-btn-text-disabled)]'

  $: isDisabled = disabled || isCooling
  $: classes = [
    base,
    sizeClasses[size],
    colorClasses[color],
    isDisabled ? disabledClasses : '',
    typeof $$props.class === 'string' ? $$props.class : '',
  ]
    .filter(Boolean)
    .join(' ')

  function handleClick(event: MouseEvent) {
    if (isDisabled) {
      event.preventDefault()
      event.stopImmediatePropagation()
      return
    }

    onclick?.(event)

    if (cooldown > 0) {
      isCooling = true
      setTimeout(() => {
        isCooling = false
      }, cooldown)
    }
  }
</script>

{#if as === 'button'}
  <button {...$$restProps} class={classes} disabled={isDisabled} onclick={handleClick}>
    <slot />
  </button>
{:else}
  <svelte:element this={as} {...$$restProps} class={classes} aria-disabled={isDisabled} onclick={handleClick}>
    <slot />
  </svelte:element>
{/if}
