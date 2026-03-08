<script lang="ts">
  import { mdiCheck, mdiClose } from '@mdi/js'

  import ZIcon from '$shared/ui/ZIcon.svelte'

  export let checked = false
  export let label = ''

  function toggle() {
    checked = !checked
  }
</script>

<button
  type="button"
  role="switch"
  aria-checked={checked}
  aria-label={label}
  class={`relative inline-flex h-6 w-11 items-center rounded-full border-2 bg-transparent transition-colors ${
    checked
      ? 'border-[var(--color-success)] text-[var(--color-success)]'
      : 'border-neutral-400 text-neutral-400 dark:border-slate-500 dark:text-slate-500'
  } cursor-pointer`}
  onclick={toggle}
>
  <span
    class={`inline-flex h-4 w-4 transform items-center justify-center rounded-full bg-current transition-transform ${checked ? 'translate-x-5.5' : 'translate-x-0.5'}`}
  >
    <ZIcon path={checked ? mdiCheck : mdiClose} size={24} class="text-(--z-base)" />
  </span>
</button>
