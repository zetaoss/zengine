<script lang="ts">
  import { mdiCheck, mdiClose } from '@mdi/js'
  import { createEventDispatcher } from 'svelte'

  import ZIcon from '$shared/ui/ZIcon.svelte'

  const dispatch = createEventDispatcher<{ change: { checked: boolean } }>()

  export let checked = false
  export let label = ''
  export let disabled = false
  export let showIcon = true
  export let theme: 'default' | 'success' = 'default'

  $: toneClass = checked
    ? theme === 'success'
      ? 'border-[var(--color-success)] text-[var(--color-success)]'
      : 'border-neutral-700 text-neutral-700 dark:border-slate-300 dark:text-slate-300'
    : 'border-neutral-400 text-neutral-400 dark:border-slate-500 dark:text-slate-500'

  function toggle() {
    if (disabled) return
    checked = !checked
    dispatch('change', { checked })
  }
</script>

<button
  type="button"
  role="switch"
  aria-checked={checked}
  aria-disabled={disabled}
  aria-label={label}
  class={`relative inline-flex h-6 w-11 items-center rounded-full border-2 bg-transparent transition-colors ${toneClass} ${
    disabled ? 'cursor-not-allowed opacity-60' : 'cursor-pointer'
  }`}
  onclick={toggle}
  {disabled}
>
  <span
    class={`inline-flex h-4 w-4 transform items-center justify-center rounded-full bg-current transition-transform ${checked ? 'translate-x-5.5' : 'translate-x-0.5'}`}
  >
    {#if showIcon}
      <ZIcon path={checked ? mdiCheck : mdiClose} size={24} class="text-(--z-base)" />
    {/if}
  </span>
</button>
