<script lang="ts">
  import ZIcon from '$shared/ui/ZIcon.svelte'

  import type { ItemData } from './types'

  export let itemData: ItemData
</script>

<button
  type="button"
  class={`menu-item ${itemData.isActive && itemData.isActive() ? 'is-active' : ''}`}
  title={itemData.title ?? ''}
  on:click={() => itemData.action?.()}
>
  <ZIcon path={itemData.icon ?? ''} />
</button>

<style lang="postcss">
  @reference 'tailwindcss';

  .menu-item {
    @apply ml-1 rounded cursor-pointer;
  }

  .menu-item :global(svg) {
    @apply h-7 w-7 p-1 fill-current;
  }

  .menu-item.is-active,
  .menu-item:hover {
    @apply bg-zinc-200 dark:bg-zinc-800;
  }
</style>
