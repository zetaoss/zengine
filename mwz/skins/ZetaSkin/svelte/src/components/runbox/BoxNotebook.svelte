<!-- BoxNotebook.svelte -->
<script lang="ts">
  import { mdiAlert } from '@mdi/js'
  import type { Writable } from 'svelte/store'

  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import { formatDateTime } from '$shared/utils/time'

  import NBOutput from './notebook/NBOutput.svelte'
  import type { Job } from './types'

  export let job: Writable<Job>
  export let seq: number
  export let wrapped = true

  $: jobValue = $job
  $: nbouts = jobValue.notebookOuts[seq] ?? []
  $: loaded = jobValue.phase != null
  $: lowerPhase = jobValue.phase?.toLowerCase()
  $: displayPhase = jobValue.isLoading ? 'loading' : lowerPhase
  $: updatedAtLabel = formatDateTime(jobValue.updatedAt)

  const containerClass = 'rounded-b-lg border-t px-4 py-2 text-sm'
</script>

<slot />

{#if displayPhase === 'loading' || displayPhase === 'pending' || displayPhase === 'running'}
  <div class="{containerClass} flex items-center">
    {#if displayPhase === 'loading' || displayPhase === 'running'}
      <ZSpinner size="0.875rem" />
    {:else if displayPhase === 'pending'}
      <ZSpinner size="0.875rem" reverse />
    {/if}
    <span class="text-muted-foreground text-xs px-1">{displayPhase}...</span>
  </div>
{/if}

{#if loaded && (lowerPhase === 'succeeded' || lowerPhase === 'failed') && nbouts.length}
  {#if lowerPhase === 'failed'}
    <div class="flex items-center gap-1 text-xs text-a-red-500 font-medium mb-2 mt-1">
      <ZIcon size={14} path={mdiAlert} />
      {#if updatedAtLabel}
        <time datetime={jobValue.updatedAt ?? undefined} title={jobValue.updatedAt ?? undefined}>
          Failed · {updatedAtLabel}
        </time>
      {:else}
        <span>Failed</span>
      {/if}
    </div>
  {/if}
  <div class={containerClass}>
    {#each nbouts as nbout, i (i)}
      <NBOutput out={nbout} {wrapped} />
    {/each}
  </div>
{/if}
