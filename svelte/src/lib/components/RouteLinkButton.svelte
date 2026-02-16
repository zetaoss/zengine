<script lang="ts">
  import { resolve } from '$app/paths'
  import type { ZButtonColor, ZButtonSize } from '$shared/ui/zButtonShared'
  import { getZButtonClasses, handleZButtonClick } from '$shared/ui/zButtonShared'

  const resolvePath = resolve as unknown as (path: string) => string

  export let to: string
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

<svelte:element this={'a'} {...$$restProps} href={resolvePath(to)} class={classes} aria-disabled={isDisabled} onclick={handleClick}>
  <slot />
</svelte:element>
