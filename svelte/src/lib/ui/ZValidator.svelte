<svelte:options runes={true} />

<script lang="ts">
  import type { Snippet } from 'svelte'

  interface Props {
    error?: string
    containerClass?: string
    left?: Snippet
    children?: Snippet<[boolean]>
  }

  let { error = '', containerClass = '', left, children }: Props = $props()
</script>

<div class="flex flex-col gap-1 {containerClass}">
  <div class="flex items-start gap-2">
    {#if left}
      {@render left()}
    {/if}
    <div class="flex flex-1 flex-col gap-1">
      {#if children}
        {@render children(error !== '')}
      {/if}
      {#if error}
        <div class="text-xs text-a-red-500 px-1">{error}</div>
      {/if}
    </div>
  </div>
</div>
