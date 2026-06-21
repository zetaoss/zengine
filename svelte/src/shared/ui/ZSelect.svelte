<svelte:options runes={true} />

<script lang="ts">
  import type { Snippet } from 'svelte'

  import { useDismissable } from '$shared/composables/useDismissable'

  interface SelectItem {
    group?: string
    value: string
    label: string
  }

  let {
    value = $bindable(''),
    items = [],
    placeholder = '선택...',
    class: className = '',
    triggerClass = '',
    onchange,
    item: itemSnippet,
  }: {
    value?: string
    items: SelectItem[]
    placeholder?: string
    class?: string
    triggerClass?: string
    onchange?: (value: string) => void
    item?: Snippet<[SelectItem]>
  } = $props()

  let open = $state(false)
  let rootEl = $state<HTMLElement | null>(null)

  let selectedItem = $derived(items.find((i) => i.value === value))

  function toggle() {
    open = !open
  }

  function select(item: SelectItem) {
    value = item.value
    open = false
    onchange?.(item.value)
  }

  useDismissable(() => rootEl, {
    enabled: () => open,
    onDismiss: () => (open = false),
  })
</script>

<div bind:this={rootEl} class="relative {className}">
  <button type="button" class="z-select flex w-full items-center justify-between gap-2 text-left {triggerClass}" onclick={toggle}>
    <div class="min-w-0 flex-1 truncate">
      {#if selectedItem}
        {#if itemSnippet}
          {@render itemSnippet(selectedItem)}
        {:else}
          {selectedItem.label}
        {/if}
      {:else}
        <span class="text-muted-foreground">{placeholder}</span>
      {/if}
    </div>
  </button>

  {#if open}
    <div
      class="bg-card ring-border absolute left-0 right-0 top-full z-50 mt-1 max-h-64 overflow-y-auto rounded border border-border py-1 shadow-lg"
    >
      {#each items as item, index (item.value)}
        {#if item.group && (index === 0 || items[index - 1].group !== item.group)}
          <div class="bg-a-slate-50 px-3 py-1 text-xs font-semibold text-a-slate-400 mt-2 first:mt-0 mb-1 border-y border-border/20">{item.group}</div>
        {/if}
        <button
          type="button"
          class="flex w-full items-center gap-2 px-3 py-1 text-left transition hover:bg-a-gray-200 {value === item.value
            ? 'bg-a-gray-100 font-semibold'
            : ''}"
          onclick={() => select(item)}
        >
          {#if itemSnippet}
            {@render itemSnippet(item)}
          {:else}
            {item.label}
          {/if}
        </button>
      {/each}
      {#if items.length === 0}
        <div class="px-3 py-2 text-sm text-muted-foreground">항목이 없습니다.</div>
      {/if}
    </div>
  {/if}
</div>
