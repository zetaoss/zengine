<script lang="ts">
  import type { ZButtonColor, ZButtonSize } from './zButtonShared'
  import { getZButtonClasses, handleZButtonClick } from './zButtonShared'

  export let href = '#'
  export let disabled = false
  export let color: ZButtonColor = 'default'
  export let size: ZButtonSize = 'medium'
  export let onclick: ((event: MouseEvent) => void) | undefined = undefined

  $: isDisabled = disabled
  $: classes = getZButtonClasses({
    size,
    color,
    isDisabled,
    className: typeof $$props.class === 'string' ? $$props.class : undefined,
  })

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
  <slot />
</a>
