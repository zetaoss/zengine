<script lang="ts">
  import type { Snippet } from 'svelte'

  import { useDismissable } from '$shared/composables/useDismissable'

  interface Props {
    disabled?: boolean
    open?: boolean
    onOpenChange?: (open: boolean) => void
    trigger?: Snippet<[{ open: boolean; toggle: () => void; close: () => void }]>
    menu?: Snippet<[{ close: () => void }]>
    children?: Snippet
  }

  let {
    disabled = false,
    open: controlledOpen,
    onOpenChange,
    trigger,
    menu,
    children,
  }: Props = $props()

  let rootEl = $state<HTMLElement | null>(null)
  let innerOpen = $state(false)

  let open = $derived(controlledOpen !== undefined ? controlledOpen : innerOpen)

  function setOpen(v: boolean) {
    if (onOpenChange) {
      onOpenChange(v)
    } else {
      innerOpen = v
    }
  }

  function close() {
    setOpen(false)
  }

  function toggle() {
    if (disabled) return
    setOpen(!open)
  }

  useDismissable(() => rootEl, {
    enabled: () => open,
    onDismiss: close,
  })
</script>

<div bind:this={rootEl} class="relative inline-flex">
  {@render trigger?.({ open, toggle, close })}

  {#if open}
    <div
      class="bg-card ring-border absolute right-0 top-full z-50 mt-1 inline-flex w-max flex-col rounded py-1 shadow-lg"
      role="menu"
      tabindex="-1"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
    >
      {@render menu?.({ close })}
      {@render children?.()}
    </div>
  {/if}
</div>
