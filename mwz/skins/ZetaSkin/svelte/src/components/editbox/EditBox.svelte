<svelte:options customElement={{ tag: 'edit-box', shadow: 'none' }} />

<script lang="ts">
  import { onMount } from 'svelte'

  import { startEditBox } from './editboxController'

  let rootElement: HTMLDivElement | null = null
  let editHeaderMountElement: HTMLDivElement | null = null

  onMount(() =>
    startEditBox({
      rootElement: rootElement as HTMLDivElement,
      editHeaderMountElement: editHeaderMountElement as HTMLDivElement,
    }),
  )
</script>

<div bind:this={rootElement}>
  <div bind:this={editHeaderMountElement}></div>
</div>
