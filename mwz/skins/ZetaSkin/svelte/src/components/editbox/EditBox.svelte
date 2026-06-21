<svelte:options customElement={{ tag: 'edit-box', shadow: 'none' }} />

<script lang="ts">
  import { onMount } from 'svelte'

  import AiEditDiff from './aiedit/AiEditDiff.svelte'
  import { startEditBox } from './editboxController'

  let rootElement: HTMLDivElement | null = null
  let editHeaderMountElement: HTMLDivElement | null = null
  let editAreaMountElement: HTMLDivElement | null = null
  let aiEditPanelMountElement: HTMLDivElement | null = null

  onMount(() =>
    startEditBox({
      aiEditPanelMountElement: aiEditPanelMountElement as HTMLDivElement,
      editAreaMountElement: editAreaMountElement as HTMLDivElement,
      rootElement: rootElement as HTMLDivElement,
      editHeaderMountElement: editHeaderMountElement as HTMLDivElement,
    }),
  )
</script>

<div bind:this={rootElement}>
  <AiEditDiff />
  <div bind:this={editHeaderMountElement}></div>
</div>
<div class="flex flex-col gap-3 md:flex-row-reverse md:items-start" bind:this={editAreaMountElement}>
  <div class="hidden min-w-0 md:flex-1" bind:this={aiEditPanelMountElement}></div>
</div>
