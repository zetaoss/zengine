<!-- BoxNotebook.svelte -->
<script lang="ts">
  import type { Writable } from 'svelte/store'

  import NBOutput from './notebook/NBOutput.svelte'
  import type { Job } from './types'

  export let job: Writable<Job>
  export let seq: number

  $: jobValue = $job
  $: nbouts = jobValue.notebookOuts[seq] ?? []
  $: loaded = jobValue.phase != null
</script>

<slot />

{#if loaded && jobValue.phase === 'succeeded' && nbouts.length}
  {#each nbouts as nbout, i (i)}
    <NBOutput out={nbout} />
  {/each}
{/if}
