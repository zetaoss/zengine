<svelte:options runes={true} />

<script lang="ts">
  import type { Snippet } from 'svelte'

  import TheTool from './TheTool.svelte'

  let { children }: { children?: Snippet } = $props()
</script>

<TheTool>
  {@render children?.()}
</TheTool>
