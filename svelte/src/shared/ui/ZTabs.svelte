<script lang="ts">
  import ZBadge from './ZBadge.svelte'

  interface ZTabItem {
    value: string
    label: string
    badge?: string | number
  }

  export let tabs: ZTabItem[] = []
  export let selected = ''
  export let onChange: ((value: string) => void) | undefined = undefined

  function select(value: string) {
    if (value === selected) return
    onChange?.(value)
  }
</script>

<div class="mb-4 inline-flex items-end">
  {#each tabs as tab (tab.value)}
    <button
      type="button"
      class={`relative border-b-2 px-3 py-2 text-sm transition ${
        selected === tab.value
          ? 'border-slate-900 font-semibold text-slate-900 dark:border-slate-100 dark:text-slate-100'
          : 'border-transparent text-gray-500 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-200'
      }`}
      onclick={() => select(tab.value)}
    >
      <span class="inline-flex items-center gap-1">
        {tab.label}
        {#if tab.badge !== undefined}
          <ZBadge text={String(tab.badge)} />
        {/if}
      </span>
    </button>
  {/each}
</div>
