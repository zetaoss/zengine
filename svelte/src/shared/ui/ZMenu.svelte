<script lang="ts">
  import { createEventDispatcher } from 'svelte'

  import { useDismissable } from '$shared/composables/useDismissable'

  export let disabled = false
  export let modelValue: boolean | undefined = undefined

  const dispatch = createEventDispatcher<{ 'update:modelValue': boolean }>()

  let rootEl: HTMLElement | null = null
  let innerOpen = false

  $: isControlled = modelValue !== undefined
  $: open = isControlled ? !!modelValue : innerOpen

  function setOpen(v: boolean) {
    if (isControlled) dispatch('update:modelValue', v)
    else innerOpen = v
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
  <slot name="trigger" {open} {toggle} {close} />

  {#if open}
    <div
      class="z-base2 z-ring absolute right-0 top-full z-50 mt-1 inline-flex w-max flex-col rounded py-1 shadow-lg"
      role="menu"
      tabindex="-1"
      on:click|stopPropagation
      on:keydown|stopPropagation
    >
      <slot name="menu" {close} />
    </div>
  {/if}
</div>
