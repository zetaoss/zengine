<script lang="ts">
  import { handleZButtonClick } from './zButtonShared'

  export let href = '#'
  export let disabled = false
  export let onclick: ((event: MouseEvent) => void) | undefined = undefined

  $: isDisabled = disabled
  $: className = typeof $$props.class === 'string' ? $$props.class : undefined
  const buttonClass =
    'inline-flex items-center justify-center rounded border px-3 py-2 text-sm leading-4 hover:no-underline transition-shadow hover:shadow-[inset_0_0_0_99em_#4447]'
  $: classes = [buttonClass, className ?? '', isDisabled ? 'pointer-events-none opacity-50' : ''].filter(Boolean).join(' ')

  function handleClick(event: MouseEvent) {
    handleZButtonClick({
      event,
      isDisabled,
      onclick,
    })
  }
</script>

<a
  {...$$restProps}
  {href}
  class={classes}
  aria-disabled={isDisabled ? 'true' : undefined}
  tabindex={isDisabled ? -1 : undefined}
  onclick={handleClick}
>
  <span class="inline-flex items-center justify-center gap-2 leading-4">
    <slot />
  </span>
</a>
