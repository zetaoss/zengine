<!-- BoxNotebook.svelte -->
<script lang="ts">
  import { mdiAlert } from '@mdi/js'
  import type { Writable } from 'svelte/store'

  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'

  import NBOutput from './notebook/NBOutput.svelte'
  import type { Job } from './types'

  export let job: Writable<Job>
  export let seq: number
  export let wrapped = true

  $: jobValue = $job
  $: nbouts = jobValue.notebookOuts[seq] ?? []
  $: loaded = jobValue.phase != null
  $: lowerPhase = jobValue.phase?.toLowerCase()

  const containerClass = 'rounded-b-lg border-t px-4 py-2 text-sm'
</script>

<slot />

{#if loaded && (lowerPhase === 'pending' || lowerPhase === 'running')}
  <div class="{containerClass} flex items-center">
    {#if lowerPhase === 'running'}
      <ZSpinner size="0.875rem" extraClass="mr-0!" />
    {:else if lowerPhase === 'pending'}
      <ZSpinner size="0.875rem" extraClass="mr-0! [animation-direction:reverse]" />
    {/if}
    <span class="text-muted-foreground text-xs">{lowerPhase}...</span>
  </div>
{/if}

{#if loaded && (lowerPhase === 'succeeded' || lowerPhase === 'failed') && nbouts.length}
  {#if lowerPhase === 'failed'}
    <div class="flex items-center text-xs text-a-red-500 font-medium mb-2 mt-1">
      <ZIcon size={14} path={mdiAlert} />
    </div>
  {/if}
  <div class={containerClass}>
    {#each nbouts as nbout, i (i)}
      <NBOutput out={nbout} {wrapped} />
    {/each}
  </div>
{/if}
