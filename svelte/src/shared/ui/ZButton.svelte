<script lang="ts">
  import type { ZButtonColor, ZButtonSize } from './zButtonShared'
  import { getZButtonClasses, handleZButtonClick } from './zButtonShared'

  export let type: 'button' | 'submit' | 'reset' = 'button'
  export let disabled = false
  export let color: ZButtonColor = 'default'
  export let size: ZButtonSize = 'medium'
  export let cooldown = 300
  export let onclick: ((event: MouseEvent) => void) | undefined = undefined

  let isCooling = false

  $: isDisabled = disabled || isCooling
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
      cooldown,
      startCooldown: (ms) => {
        isCooling = true
        setTimeout(() => {
          isCooling = false
        }, ms)
      },
    })
  }
</script>

<button {...$$restProps} {type} class={classes} disabled={isDisabled} onclick={handleClick}>
  <slot />
</button>
