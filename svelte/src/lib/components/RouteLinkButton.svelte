<script lang="ts">
  import { resolve } from '$app/paths'
  import CButton, { type ButtonSize, type ButtonVariant } from '$shared/ui/CButton.svelte'

  export let to: string
  export let disabled = false
  export let variant: ButtonVariant = 'outline'
  export let size: ButtonSize = 'medium'
  export let onclick: ((event: MouseEvent) => void) | undefined = undefined

  const resolvePath = resolve as unknown as (path: string) => string
</script>

<CButton
  {...$$restProps}
  class={typeof $$props.class === 'string' ? $$props.class : undefined}
  href={resolvePath(to)}
  {disabled}
  {variant}
  {size}
  {onclick}
>
  <slot />
</CButton>
