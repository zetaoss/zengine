<script lang="ts">
  import { mdiCheck, mdiClose } from '@mdi/js'

  import ZIcon from '$shared/ui/ZIcon.svelte'

  export let checked = false
  export let label = ''
  export let disabled = false
  export let showIcon = false
  export let onchange: ((event: { checked: boolean }) => void) | undefined = undefined

  function toggle() {
    if (disabled) return
    checked = !checked
    onchange?.({ checked })
  }
</script>

<button
  type="button"
  role="switch"
  aria-checked={checked}
  aria-disabled={disabled}
  aria-label={label}
  class="z-toggle relative inline-flex h-5 w-9 items-center rounded-full border-2 bg-transparent transition"
  class:cursor-not-allowed={disabled}
  class:opacity-40={!checked}
  class:opacity-60={disabled && checked}
  onclick={toggle}
  {disabled}
>
  <span
    class={`inline-flex h-3 w-3 transform items-center justify-center rounded-full bg-current transition-transform ${checked ? 'translate-x-4.5' : 'translate-x-1'}`}
  >
    {#if showIcon}
      <ZIcon path={checked ? mdiCheck : mdiClose} size={24} />
    {/if}
  </span>
</button>

<style>
  .z-toggle {
    border-color: var(--color-subtle);
    color: var(--color-subtle);
  }
</style>
